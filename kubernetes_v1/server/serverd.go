package main

import (
	"context"
	"flag"
	"html/template"
	"kubernetes/app"
	"kubernetes/auth"
	"kubernetes/conf"
	"kubernetes/db"
	"kubernetes/k8s"
	"kubernetes/user"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func testUserHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("template/user.html")
	if err != nil {
		w.Write([]byte("parse template error: " + err.Error()))
		return
	}
	t.Execute(w, nil)
}

func testInstanceHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("template/spark.html")
	if err != nil {
		w.Write([]byte("parse template error: " + err.Error()))
		return
	}
	t.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("template/login.html")
	if err != nil {
		w.Write([]byte("parse template error: " + err.Error()))
		return
	}
	t.Execute(w, nil)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "kubernetes_token", Expires: time.Now().Add(time.Hour * -24)})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"logout successful"}`))
}

func main() {
	configFile := flag.String("config", "server/server.ini", "config file for server")
	flag.Parse()
	myConfig := new(conf.Config)
	myConfig.InitConfig(configFile)
	logFileName := myConfig.Read("logfile")
	if logFileName == "" {
		logFileName = "server/server.log"
	}
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Panicln("failed to open log file: " + logFileName)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	app.FirstProxyAddr = myConfig.Read("firstproxyaddr")
	if app.FirstProxyAddr == "" {
		app.FirstProxyAddr = "http://0.0.0.0:9090/proxy"
	}
	app.SecondProxyAddr = myConfig.Read("secondproxyaddr")
	if app.SecondProxyAddr == "" {
		app.SecondProxyAddr = "http://0.0.0.0:9999/proxy"
	}
	ip := myConfig.Read("ip")
	if ip == "" {
		ip = "0.0.0.0"
	}
	app.SecondProxyIp = ip
	port := myConfig.Read("port")
	if port == "" {
		port = "8080"
	}
	auth.LdapServer = myConfig.Read("ldapserver")
	if auth.LdapServer == "" {
		auth.LdapServer = "ldap://10.127.48.18/"
	}
	auth.BindDN = myConfig.Read("binddn")
	if auth.BindDN == "" {
		auth.BindDN = "cn=root,dc=nscc,dc=com"
	}
	auth.BindPassword = myConfig.Read("bindpassword")
	if auth.BindPassword == "" {
		auth.BindPassword = "bigdata-admin"
	}
	auth.UserDN = myConfig.Read("userdn")
	if auth.UserDN == "" {
		auth.UserDN = "uid=%s,ou=People,dc=nscc,dc=com"
	}
	auth.GroupDN = myConfig.Read("groupdn")
	if auth.GroupDN == "" {
		auth.GroupDN = "cn=%s,ou=Groups,dc=nscc,dc=com"
	}
	privateKeyPath := myConfig.Read("privatekey")
	if privateKeyPath == "" {
		privateKeyPath = "auth/kubernetes.rsa"
	}
	publicKeyPath := myConfig.Read("publickey")
	if publicKeyPath == "" {
		publicKeyPath = "auth/kubernetes.rsa.pub"
	}
	auth.JwtInit(privateKeyPath, publicKeyPath)
	dbstr := myConfig.Read("dbstr")
	if dbstr == "" {
		dbstr = "kubernetes:kubernetes@(10.127.48.18:3306)/kubernetes"
	}
	db.InitMysqlDB(dbstr)
	k8s.BearerToken = myConfig.Read("bearertoken")
	if k8s.BearerToken == "" {
		k8s.BearerToken = "Bearer 6B1GbqhcjqGYPAAy285otYhUUV4z4kiu"
	}
	k8s.K8sApiServer = myConfig.Read("k8sapiserver")
	if k8s.K8sApiServer == "" {
		k8s.K8sApiServer = "https://10.127.48.18:6443/api/v1"
	}
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/testuser", testUserHandler)
	serveMux.HandleFunc("/testinstance", testInstanceHandler)
	serveMux.HandleFunc("/login", loginHandler)
	serveMux.HandleFunc("/logout", logoutHandler)
	serveMux.HandleFunc("/img/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "template/"+r.URL.Path[1:])
	})
	serveMux.HandleFunc("/js/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "template/"+r.URL.Path[1:])
	})
	serveMux.HandleFunc("/css/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "template/"+r.URL.Path[1:])
	})
	serveMux.HandleFunc("/auth", auth.AuthHandler)
	serveMux.HandleFunc("/user", user.UserHandler)
	serveMux.HandleFunc("/group", user.GroupHandler)
	serveMux.HandleFunc("/app", app.AppHandler)
	serveMux.HandleFunc("/instance", app.InstanceHandler)
	server := &http.Server{Addr: ip + ":" + port, Handler: serveMux}
	log.Println("starting Server ...")
	log.Println("listening on address " + ip + ":" + port)
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ListenAndServe()
		log.Println("shutting down server ...")
	}()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	ctx := context.Background()
	for {
		select {
		case err := <-errChan:
			if err != nil {
				log.Println(err)
				db.CloseMysqlDB()
				return
			}
		case s := <-signalChan:
			log.Printf("Captured %v. Exiting ...\n", s)
			server.Shutdown(ctx)
		}
	}
}
