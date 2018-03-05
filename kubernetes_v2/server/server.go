package main

import (
	"context"
	htmlTemplate "html/template"
	"kubernetes/app"
	"kubernetes/auth"
	"kubernetes/conf"
	"kubernetes/db"
	"kubernetes/group"
	"kubernetes/k8s"
	"kubernetes/template"
	"kubernetes/user"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func testUserHandler(w http.ResponseWriter, r *http.Request) {
	t, err := htmlTemplate.ParseFiles("html/user.html")
	if err != nil {
		log.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal server error!"}`))
		return
	}
	t.Execute(w, nil)
}

func testAppHandler(w http.ResponseWriter, r *http.Request) {
	var (
		t    *htmlTemplate.Template
		name string
		err  error
	)
	r.ParseForm()
	name = r.FormValue("name")
	switch {
	case name == "spark":
		t, err = htmlTemplate.ParseFiles("html/spark.html")
	case name == "slurm":
		t, err = htmlTemplate.ParseFiles("html/slurm.html")
	case name == "tensorflow":
		t, err = htmlTemplate.ParseFiles("html/tensorflow.html")
	case name == "rstudio":
		t, err = htmlTemplate.ParseFiles("html/rstudio.html")
	default:
		log.Println("no such application!")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"no such application!"}`))
		return
	}
	if err != nil {
		log.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal server error!"}`))
		return
	}
	t.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	t, err := htmlTemplate.ParseFiles("html/login.html")
	if err != nil {
		log.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal server error!"}`))
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
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	config := new(conf.Config)
	err := config.InitConfig("server/server.ini")
	if err != nil {
		log.Fatalln(err)
	}
	logFileName := config.Get("logFile")
	if logFileName == "" {
		log.Fatalln("logFile must be set")
	}
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln("failed to open log file: " + logFileName)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	ip := config.Get("ip")
	if ip == "" {
		log.Fatalln("ip must be set")
	}
	port := config.Get("port")
	if port == "" {
		log.Fatalln("prot must be set")
	}
	app.ProxyAddr = config.Get("proxyAddr")
	if app.ProxyAddr == "" {
		log.Fatalln("proxyAddr must be set")
	}
	auth.LDAPServer = config.Get("LDAPServer")
	if auth.LDAPServer == "" {
		log.Fatalln("LDAPServer must be set")
	}
	auth.BindDN = config.Get("bindDN")
	if auth.BindDN == "" {
		log.Fatalln("bindDN must be set")
	}
	auth.BindPassWord = config.Get("bindPassWord")
	if auth.BindPassWord == "" {
		log.Fatalln("bindPassWord must be set")
	}
	auth.UserDN = config.Get("userDN")
	if auth.UserDN == "" {
		log.Fatalln("userDN must be set")
	}
	auth.GroupDN = config.Get("groupDN")
	if auth.GroupDN == "" {
		log.Fatalln("groupDN must be set")
	}
	privateKeyPath := config.Get("privateKey")
	if privateKeyPath == "" {
		log.Fatalln("privateKey must be set")
	}
	publicKeyPath := config.Get("publicKey")
	if publicKeyPath == "" {
		log.Fatalln("publicKey must be set")
	}
	err = auth.JwtInit(privateKeyPath, publicKeyPath)
	if err != nil {
		log.Fatalln(err)
	}
	dbStr := config.Get("dbStr")
	if dbStr == "" {
		log.Fatalln("dbStr must be set")
	}
	err = db.InitMysqlDB(dbStr)
	if err != nil {
		log.Fatalln(err)
	}
	err = k8s.InitK8s()
	if err != nil {
		log.Fatalln(err)
	}
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/testuser", testUserHandler)
	serveMux.HandleFunc("/testapp", testAppHandler)
	serveMux.HandleFunc("/login", loginHandler)
	serveMux.HandleFunc("/logout", logoutHandler)
	serveMux.HandleFunc("/img/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "html/"+r.URL.Path[1:])
	})
	serveMux.HandleFunc("/js/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "html/"+r.URL.Path[1:])
	})
	serveMux.HandleFunc("/css/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "html/"+r.URL.Path[1:])
	})
	serveMux.HandleFunc("/auth", auth.AuthHandler)
	serveMux.HandleFunc("/user", user.UserHandler)
	serveMux.HandleFunc("/group", group.GroupHandler)
	serveMux.HandleFunc("/template", template.TemplateHandler)
	serveMux.HandleFunc("/app", app.AppHandler)
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
				err = db.CloseMysqlDB()
				if err != nil {
					log.Fatalln(err)
				}
				return
			}
		case s := <-signalChan:
			log.Println("Captured %v. Exiting ...", s)
			server.Shutdown(ctx)
		}
	}
}
