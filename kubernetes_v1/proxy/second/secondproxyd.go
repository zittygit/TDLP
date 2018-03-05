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
	proxyServerMap map[int]*ProxyServerList
	proxyRWMutex   *sync.RWMutex
	firstProxyIP   string
	secondProxyIP  string
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
		webSocketUrl string
		port  int
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
	rows, err := db.MysqlDB.Query("select firstport, websocketurl from proxy where pid=" + strconv.Itoa(proxy.Pid))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query proxy info"}`))
		return
	}
	if rows.Next() {
		err = rows.Scan(&port, &webSocketUrl)
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
	proxyServer.Listener, err = net.Listen("tcp", secondProxyIP+":0")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotModified)
		w.Write([]byte(`{"message":"can not create proxy for application"}`))
		return
	}
	proxyServer.Port = proxyServer.Listener.Addr().(*net.TCPAddr).Port
	remote, err := url.Parse("http://" + firstProxyIP + ":" + strconv.Itoa(port))
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
		remote, err = url.Parse("ws://" + firstProxyIP + ":" + strconv.Itoa(port))
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
	_, err = db.MysqlDB.Exec("update proxy set secondport=" + strconv.Itoa(proxyServer.Port) + " where pid=" + strconv.Itoa(proxy.Pid))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to update secondport in mysql"}`))
		return
	}
	server := &http.Server{Handler: serveMux}
	go func() {
		proxyServer.errChan = make(chan error, 1)
		proxyServer.errChan <- server.Serve(proxyServer.Listener)
	}()
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message":"proxy for this application created successful"}`))
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

func main() {
	proxyRWMutex = new(sync.RWMutex)
	proxyServerMap = make(map[int]*ProxyServerList)
	configFile := flag.String("config", "proxy/second/secondproxy.ini", "config file for second proxy server")
	flag.Parse()
	myConfig := new(conf.Config)
	myConfig.InitConfig(configFile)
	logFileName := myConfig.Read("logfile")
	if logFileName == "" {
		logFileName = "proxy/second/secondproxy.log"
	}
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Println(err)
		log.Panicln("failed to open log file: " + logFileName)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	firstProxyIP = myConfig.Read("firstproxyip")
	if firstProxyIP == "" {
		firstProxyIP = "0.0.0.0"
	}
	firstProxyPort := myConfig.Read("firstproxyport")
	if firstProxyPort == "" {
		firstProxyPort = "9090"
	}
	secondProxyIP = myConfig.Read("secondproxyip")
	if secondProxyIP == "" {
		secondProxyIP = "0.0.0.0"
	}
	secondProxyPort := myConfig.Read("secondproxyport")
	if secondProxyPort == "" {
		secondProxyPort = "9999"
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
	token, err := auth.JwtCreateToken(0, "guoguixin", "admin")
	if err != nil {
		log.Println(err)
		return
	}
	req, err := http.NewRequest("POST", "http://" + firstProxyIP + ":" + firstProxyPort + "/state", nil)
	if err != nil {
		log.Println(err)
		return
	}
	req.AddCookie(&http.Cookie{Name: "kubernetes_token", Value: token, Expires: time.Now().Add(time.Hour * 24)})
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	if res.StatusCode == http.StatusCreated {
		data, _ := ioutil.ReadAll(res.Body)
        	res.Body.Close()
		log.Println(string(data))
	} else {
		log.Println("Please start the first proxy server first!")
		return
	}
	rows, err := db.MysqlDB.Query("select uid, pid, firstport, websocketurl from instance, proxy where instance.iid=proxy.iid")
	if err != nil {
		log.Println(err)
		db.CloseMysqlDB()
		return
	}
	var (
		uid int
		pid int
		firstPort string
		webSocketUrl string
	)
	for rows.Next() {
		err = rows.Scan(&uid, &pid, &firstPort, &webSocketUrl)
		if err != nil {
			log.Println(err)
	                db.CloseMysqlDB()
        	        return
		}
		serveMux := http.NewServeMux()
		proxyServer := new(ProxyServer)
		proxyServer.Uid = uid
		proxyServer.Listener, err = net.Listen("tcp", secondProxyIP+":0")
		if err != nil {
			log.Println(err)
			continue
		}
		proxyServer.Port = proxyServer.Listener.Addr().(*net.TCPAddr).Port
		remote, err := url.Parse("http://" + firstProxyIP + ":" + firstPort)
		if err != nil {
			log.Println(err)
			continue
		}
		proxyServer.HttpProxy = httputil.NewSingleHostReverseProxy(remote)
		serveMux.HandleFunc("/", proxyServer.HttpProxyHandler)
		if webSocketUrl != "" {
			remote, err = url.Parse("ws://" + firstProxyIP + ":" + firstPort)
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
		_, err = db.MysqlDB.Exec("update proxy set secondport=" + strconv.Itoa(proxyServer.Port) + " where pid=" + strconv.Itoa(pid))
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
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/proxy", proxyHandler)
	server := &http.Server{Addr: secondProxyIP + ":" + secondProxyPort, Handler: httpMux}
	log.Println("Starting second Proxy Server ...")
	log.Println("Listening on address " + secondProxyIP + ":" + secondProxyPort)
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ListenAndServe()
		log.Println("shutting down second proxy server ...")
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
