package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/koding/websocketproxy"
	"io/ioutil"
	"kubernetes/auth"
	"kubernetes/conf"
	"kubernetes/db"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type Proxy struct {
	Pid          int `json:"pid"`
}

type ProxyServer struct {
	Uid            int
	HttpProxy      *httputil.ReverseProxy
	WebSocketProxy *websocketproxy.WebsocketProxy
	Port           int
	Listener       net.Listener
	errChan        chan error
}

type ProxyServerList struct {
	ProxyServerList map[int]*ProxyServer
	ProxyRWMutex    *sync.RWMutex
}

var (
	started        bool
	startRWMutex   *sync.RWMutex
	httpMux        *http.ServeMux
	proxyServerMap map[int]*ProxyServerList
	proxyRWMutex   *sync.RWMutex
	ip             string
)

func (proxyServer *ProxyServer) HttpProxyHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.JwtAuthRequest(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Access Unauthorized!"}`))
		return
	}
	claim := token.Claims.(*auth.KubernetesClaims)
	if claim.Uid != proxyServer.Uid {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"You Can not Access Other's Application!"}`))
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	proxyServer.HttpProxy.ServeHTTP(w, r)
}

func (proxyServer *ProxyServer) WebSocketProxyHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.JwtAuthRequest(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Access Unauthorized!"}`))
		return
	}
	claim := token.Claims.(*auth.KubernetesClaims)
	if claim.Uid != proxyServer.Uid {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"You Can not Access Other's Application!"}`))
		return
	}
	r.Header.Set("Host", r.Header.Get("Origin")[7:])
	proxyServer.WebSocketProxy.ServeHTTP(w, r)
}

func addProxyServer(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var (
		proxy Proxy
		httpUrl string
		webSocketUrl string
	)
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error"}`))
		return
	}
	if proxy.Pid == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"pid is required"}`))
		return
	}
	rows, err := db.MysqlDB.Query("select httpurl, websocketurl from proxy where pid=" + strconv.Itoa(proxy.Pid))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query proxy info"}`))
		return
	}
	if rows.Next() {
		err = rows.Scan(&httpUrl, &webSocketUrl)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query proxy info"}`))
			return
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query proxy info"}`))
		return
	}
	serveMux := http.NewServeMux()
	proxyServer := new(ProxyServer)
	proxyServer.Uid = claim.Uid
	proxyServer.Listener, err = net.Listen("tcp", ip+":0")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotModified)
		w.Write([]byte(`{"message":"can not create proxy for application"}`))
		return
	}
	proxyServer.Port = proxyServer.Listener.Addr().(*net.TCPAddr).Port
	remote, err := url.Parse(httpUrl)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"http url format error"}`))
		proxyServer.Listener.Close()
		return
	}
	proxyServer.HttpProxy = httputil.NewSingleHostReverseProxy(remote)
	serveMux.HandleFunc("/", proxyServer.HttpProxyHandler)
	if webSocketUrl != "" {
		remote, err = url.Parse(webSocketUrl)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"websocket url format error"}`))
			proxyServer.Listener.Close()
			return
		}
		proxyServer.WebSocketProxy = websocketproxy.NewProxy(remote)
		proxyServer.WebSocketProxy.Director = func(req *http.Request, requsetHeader http.Header) {
			requsetHeader.Add("Host", req.Header.Get("Host"))
		}
		serveMux.HandleFunc("/ws", proxyServer.WebSocketProxyHandler)
	}
	proxyRWMutex.Lock()
	proxyServerList, found := proxyServerMap[claim.Uid]
	if !found {
		proxyServerList = new(ProxyServerList)
		proxyServerList.ProxyServerList = make(map[int]*ProxyServer)
		proxyServerList.ProxyRWMutex = new(sync.RWMutex)
		proxyServerList.ProxyServerList[proxy.Pid] = proxyServer
		proxyServerMap[claim.Uid] = proxyServerList
		proxyRWMutex.Unlock()
	} else {
		proxyRWMutex.Unlock()
		proxyServerList.ProxyRWMutex.Lock()
		_, found := proxyServerList.ProxyServerList[proxy.Pid]
		if !found {
			proxyServerList.ProxyServerList[proxy.Pid] = proxyServer
			proxyServerList.ProxyRWMutex.Unlock()
		} else {
			proxyServerList.ProxyRWMutex.Unlock()
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"proxy for this application already exists"}`))
			proxyServer.Listener.Close()
			return
		}
	}
	_, err = db.MysqlDB.Exec("update proxy set firstport=" + strconv.Itoa(proxyServer.Port) + " where pid=" + strconv.Itoa(proxy.Pid))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to update firstport in mysql"}`))
		return
	}
	server := &http.Server{Handler: serveMux}
	go func() {
		proxyServer.errChan = make(chan error, 1)
		proxyServer.errChan <- server.Serve(proxyServer.Listener)
	}()
	w.WriteHeader(http.StatusCreated)
}

func deleteProxyServer(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var proxy Proxy
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error"}`))
		return
	}
	if proxy.Pid == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"pid is required"}`))
		return
	}
	proxyRWMutex.RLock()
	proxyServerList, found := proxyServerMap[claim.Uid]
	proxyRWMutex.RUnlock()
	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"proxy for this application not found"}`))
	} else {
		proxyServerList.ProxyRWMutex.Lock()
		proxyServer, found := proxyServerList.ProxyServerList[proxy.Pid]
		if !found {
			proxyServerList.ProxyRWMutex.Unlock()
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"proxy for this application not found"}`))
		} else {
			delete(proxyServerList.ProxyServerList, proxy.Pid)
			proxyServerList.ProxyRWMutex.Unlock()
			proxyServer.Listener.Close()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"proxy for this application delete successful"}`))
		}
	}
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	token, err := auth.JwtAuthRequest(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Access Unauthorized!"}`))
		return
	}
	claim := token.Claims.(*auth.KubernetesClaims)
	if r.Method == "POST" {
		addProxyServer(w, r, claim)
		return
	}
	if r.Method == "DELETE" {
		deleteProxyServer(w, r, claim)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"message":"only method POST and DELETE are supported"}`))
}

func stateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	token, err := auth.JwtAuthRequest(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Access Unauthorized!"}`))
		return
	}
	claim := token.Claims.(*auth.KubernetesClaims)
	if claim.Role != "admin" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Only Admin Can Use This Interface!"}`))
		return
	}
	if r.Method == "POST" {
		startRWMutex.Lock()
		if !started {
			httpMux.HandleFunc("/proxy", proxyHandler)
			started = true
		}
		startRWMutex.Unlock()
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"state":` + strconv.FormatBool(started) + "}"))
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
        w.Write([]byte(`{"message":"only method POST is supported"}`))
}

func main() {
	started = false
	startRWMutex = new(sync.RWMutex)
	proxyRWMutex = new(sync.RWMutex)
	proxyServerMap = make(map[int]*ProxyServerList)
	configFile := flag.String("config", "proxy/first/firstproxy.ini", "config file for first proxy server")
	flag.Parse()
	myConfig := new(conf.Config)
	myConfig.InitConfig(configFile)
	logFileName := myConfig.Read("logfile")
	if logFileName == "" {
		logFileName = "proxy/first/firstproxy.log"
	}
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Println(err)
		log.Panicln("failed to open log file: " + logFileName)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	ip = myConfig.Read("ip")
	if ip == "" {
		ip = "0.0.0.0"
	}
	port := myConfig.Read("port")
	if port == "" {
		port = "9090"
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
	rows, err := db.MysqlDB.Query("select uid, pid, httpurl, websocketurl from instance, proxy where instance.iid=proxy.iid")
	if err != nil {
		log.Println(err)
		db.CloseMysqlDB()
		return
	}
	var (
		uid int
		pid int
		httpUrl string
		webSocketUrl string
	)
	for rows.Next() {
		err = rows.Scan(&uid, &pid, &httpUrl, &webSocketUrl)
		if err != nil {
			log.Println(err)
	                db.CloseMysqlDB()
        	        return
		}
		serveMux := http.NewServeMux()
		proxyServer := new(ProxyServer)
		proxyServer.Uid = uid
		proxyServer.Listener, err = net.Listen("tcp", ip+":0")
		if err != nil {
			log.Println(err)
			continue
		}
		proxyServer.Port = proxyServer.Listener.Addr().(*net.TCPAddr).Port
		remote, err := url.Parse(httpUrl)
		if err != nil {
			log.Println(err)
			continue
		}
		proxyServer.HttpProxy = httputil.NewSingleHostReverseProxy(remote)
		serveMux.HandleFunc("/", proxyServer.HttpProxyHandler)
		if webSocketUrl != "" {
			remote, err = url.Parse(webSocketUrl)
			if err != nil {
				log.Println(err)
				continue
			}
			proxyServer.WebSocketProxy = websocketproxy.NewProxy(remote)
			proxyServer.WebSocketProxy.Director = func(req *http.Request, requsetHeader http.Header) {
				requsetHeader.Add("Host", req.Header.Get("Host"))
			}
			serveMux.HandleFunc("/ws", proxyServer.WebSocketProxyHandler)
		}
		proxyServerList, found := proxyServerMap[uid]
		if !found {
			proxyServerList = new(ProxyServerList)
			proxyServerList.ProxyRWMutex = new(sync.RWMutex)
			proxyServerList.ProxyServerList = make(map[int]*ProxyServer)
			proxyServerList.ProxyServerList[pid] = proxyServer
			proxyServerMap[uid] = proxyServerList
		} else {
			proxyServerList.ProxyServerList[pid] = proxyServer
		}
		_, err = db.MysqlDB.Exec("update proxy set firstport=" + strconv.Itoa(proxyServer.Port) + " where pid=" + strconv.Itoa(pid))
		if err != nil {
			log.Println(err)
			continue
		}
		server := &http.Server{Handler: serveMux}
		go func() {
			proxyServer.errChan = make(chan error, 1)
			proxyServer.errChan <- server.Serve(proxyServer.Listener)
		}()
	}
	httpMux = http.NewServeMux()
	httpMux.HandleFunc("/state", stateHandler)
	server := &http.Server{Addr: ip + ":" + port, Handler: httpMux}
	log.Println("Starting first Proxy Server ...")
	log.Println("Listening on address " + ip + ":" + port)
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ListenAndServe()
		log.Println("shutting down first proxy server ...")
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
				for _, list := range proxyServerMap {
					for _, v := range list.ProxyServerList {
						v.Listener.Close()
					}
				}
				for _, list := range proxyServerMap {
					for _, v := range list.ProxyServerList {
						for {
							err = <-v.errChan
							if err != nil {
								log.Println(err)
								break
							}
							time.Sleep(time.Second)
						}
					}
				}
				return
			}
		case s := <-signalChan:
			log.Printf("Captured %v. Exiting ...\n", s)
			server.Shutdown(ctx)
		}
	}
}
