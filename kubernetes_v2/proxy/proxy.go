package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/koding/websocketproxy"
	"io/ioutil"
	"auth"
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
	"strings"
	"sync"
	"syscall"
	"time"
)

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

type Proxy struct {
	Pid       int    `json:"pid"`
	ProxyName string `json:"proxyName"`
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	Protocol  string `json:"protocol"`
	Suffix    string `json:"suffix"`
	WSSuffix  string `json:"wsSuffix"`
}

type ProxyServer struct {
	Uid            int
	ProxyName      string
	Port           int
	Protocol       string
	Suffix         string
	WSSuffix       string
	HttpProxy      *httputil.ReverseProxy
	HttpsProxy     *httputil.ReverseProxy
	WebSocketProxy *websocketproxy.WebsocketProxy
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
	ip             string
	upstream       string
	proxyAddr      string
	certFile       string
	keyFile        string
	tp             *http.Transport
)

func (proxyServer *ProxyServer) VerifyAuthorization(w http.ResponseWriter, r *http.Request) bool {
	token, err := auth.JwtAuthRequest(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Access Unauthorized!"}`))
		return false
	}
	claim := token.Claims.(*auth.KubernetesClaims)
	if claim.Uid != proxyServer.Uid {
		log.Println("User " + claim.UserName + " intend to access other's application!")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"You Can Not Access Other's Application!"}`))
		return false
	}
	return true
}

func (proxyServer *ProxyServer) HttpProxyHandler(w http.ResponseWriter, r *http.Request) {
	if proxyServer.VerifyAuthorization(w, r) {
		w.Header().Set("Cache-Control", "no-cache")
		proxyServer.HttpProxy.ServeHTTP(w, r)
	}
}

func (proxyServer *ProxyServer) HttpsProxyHandler(w http.ResponseWriter, r *http.Request) {
	if proxyServer.VerifyAuthorization(w, r) {
		w.Header().Set("Cache-Control", "no-cache")
		proxyServer.HttpsProxy.ServeHTTP(w, r)
	}
}

func (proxyServer *ProxyServer) WebSocketProxyHandler(w http.ResponseWriter, r *http.Request) {
	if proxyServer.VerifyAuthorization(w, r) {
		r.Header.Set("Host", r.Header.Get("Origin")[7:])
		proxyServer.WebSocketProxy.ServeHTTP(w, r)
	}
}

func queryProxyServer(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	r.ParseForm()
	pid, err := strconv.Atoi(r.FormValue("pid"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"pid is required and must be integer!"}`))
		return
	}
	proxyRWMutex.RLock()
	proxyServerList, found := proxyServerMap[claim.Uid]
	proxyRWMutex.RUnlock()
	if !found {
		log.Println("proxys for user " + claim.UserName + " not found!")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"proxys for user ` + claim.UserName + ` not found!"}`))
	} else {
		proxyServerList.ProxyRWMutex.RLock()
		proxyServer, found := proxyServerList.ProxyServerList[pid]
		if !found {
			proxyServerList.ProxyRWMutex.RUnlock()
			log.Println("proxys with pid " + strconv.Itoa(pid) + " not found!")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"proxys with pid ` + strconv.Itoa(pid) + ` not found!"}`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"proxyName":"` + proxyServer.ProxyName + `","ip":"` + ip + `","port":` + strconv.Itoa(proxyServer.Port) + `,"protocol":"` + proxyServer.Protocol + `","suffix":"` + proxyServer.Suffix + `","wsSuffix":"` + proxyServer.WSSuffix + `"}`))
			proxyServerList.ProxyRWMutex.RUnlock()
		}
	}
}

func createProxyServer(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	var (
		proxy     Proxy
		proxyName string
		destIP    string
		port      int
		protocol  string
		suffix    string
		wsSuffix  string
	)
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error!"}`))
		return
	}
	if proxy.Pid == 0 {
		log.Println("pid is required!")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"pid is required!"}`))
		return
	}
	if upstream == "" {
		proxys, err := db.MysqlDB.Query("select proxyName, ip, port, protocol, suffix, wsSuffix from proxys where pid=" + strconv.Itoa(proxy.Pid))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query proxy info!"}`))
			return
		}
		defer proxys.Close()
		if proxys.Next() {
			err = proxys.Scan(&proxyName, &destIP, &port, &protocol, &suffix, &wsSuffix)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query proxy info!"}`))
				return
			}
		} else {
			log.Println("proxy with pid " + strconv.Itoa(proxy.Pid) + " and uid " + strconv.Itoa(claim.Uid) + " not found!")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"proxy with pid ` + strconv.Itoa(proxy.Pid) + " and uid " + strconv.Itoa(claim.Uid) + ` not found!"}`))
			return
		}
	} else {
		token, err := r.Cookie("kubernetes_token")
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"internal server error!"}`))
			return
		}
		req, err := http.NewRequest("POST", upstream+"/proxy", strings.NewReader(`{"pid":`+strconv.Itoa(proxy.Pid)+`}`))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"internal server error!"}`))
			return
		}
		req.AddCookie(token)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"internal server error!"}`))
			return
		}
		data, _ = ioutil.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode != http.StatusCreated {
			log.Println(string(data))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"internal server error!"}`))
			return
		}
		err = json.Unmarshal(data, &proxy)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"internal server error!"}`))
			return
		}
		proxyName, destIP, port, protocol, suffix, wsSuffix = proxy.ProxyName, proxy.IP, proxy.Port, proxy.Protocol, proxy.Suffix, proxy.WSSuffix
	}
	serveMux := http.NewServeMux()
	proxyServer := new(ProxyServer)
	proxyServer.Uid = claim.Uid
	proxyServer.Listener, err = net.Listen("tcp", ip+":0")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal server error!"}`))
		return
	}
	proxyServer.ProxyName = proxyName
	proxyServer.Protocol = protocol
	proxyServer.Suffix = suffix
	proxyServer.WSSuffix = wsSuffix
	proxyServer.Port = proxyServer.Listener.Addr().(*net.TCPAddr).Port
	if protocol == "http" {
		remote, err := url.Parse("http://" + destIP + ":" + strconv.Itoa(port))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"internal server error!"}`))
			return
		}
		proxyServer.HttpProxy = httputil.NewSingleHostReverseProxy(remote)
		serveMux.HandleFunc("/", proxyServer.HttpProxyHandler)
	} else {
		remote, err := url.Parse("https://" + destIP + ":" + strconv.Itoa(port))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"internal server error!"}`))
			return
		}
		proxyServer.HttpsProxy = httputil.NewSingleHostReverseProxy(remote)
		proxyServer.HttpsProxy.Transport = tp
		serveMux.HandleFunc("/", proxyServer.HttpsProxyHandler)
	}
	if wsSuffix != "" {
		remote, err := url.Parse("ws://" + destIP + ":" + strconv.Itoa(port) + "/" + suffix)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"internal server error!"}`))
			return
		}
		proxyServer.WebSocketProxy = websocketproxy.NewProxy(remote)
		proxyServer.WebSocketProxy.Director = func(req *http.Request, requsetHeader http.Header) {
			requsetHeader.Add("Host", req.Header.Get("Host"))
		}
		serveMux.HandleFunc("/"+wsSuffix, proxyServer.WebSocketProxyHandler)
	}
	proxyRWMutex.Lock()
	proxyServerList, found := proxyServerMap[claim.Uid]
	if !found {
		proxyServerList = new(ProxyServerList)
		proxyServerList.ProxyRWMutex = new(sync.RWMutex)
		proxyServerList.ProxyServerList = make(map[int]*ProxyServer)
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
			log.Println("proxy with pid " + strconv.Itoa(proxy.Pid) + " already exists!")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"proxy with pid ` + strconv.Itoa(proxy.Pid) + ` already exists!"}`))
			proxyServer.Listener.Close()
			return
		}
	}
	server := &http.Server{Handler: serveMux}
	go func() {
		proxyServer.errChan = make(chan error, 1)
		if protocol == "http" {
			proxyServer.errChan <- server.Serve(tcpKeepAliveListener{proxyServer.Listener.(*net.TCPListener)})
		} else {
			proxyServer.errChan <- server.ServeTLS(tcpKeepAliveListener{proxyServer.Listener.(*net.TCPListener)}, certFile, keyFile)
		}
	}()
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"proxyName":"` + proxyServer.ProxyName + `","ip":"` + ip + `","port":` + strconv.Itoa(proxyServer.Port) + `,"protocol":"` + protocol + `","suffix":"` + suffix + `","wsSuffix":"` + wsSuffix + `"}`))
}

func deleteProxyServer(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	var proxy Proxy
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	err := json.Unmarshal(data, &proxy)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error!"}`))
		return
	}
	if proxy.Pid == 0 {
		log.Println("pid is required!")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"pid is required!"}`))
		return
	}
	if upstream != "" {
		token, err := r.Cookie("kubernetes_token")
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"internal server error!"}`))
			return
		}
		req, err := http.NewRequest("DELETE", upstream+"/proxy", strings.NewReader(`{"pid":`+strconv.Itoa(proxy.Pid)+`}`))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"internal server error!"}`))
			return
		}
		req.AddCookie(token)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"internal server error!"}`))
			return
		}
		data, _ = ioutil.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode != http.StatusOK {
			log.Println(string(data))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"internal server error!"}`))
			return
		}
	}
	proxyRWMutex.RLock()
	proxyServerList, found := proxyServerMap[claim.Uid]
	proxyRWMutex.RUnlock()
	if !found {
		log.Println("proxy with pid " + strconv.Itoa(proxy.Pid) + " not found!")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"proxy with pid ` + strconv.Itoa(proxy.Pid) + ` not found!"}`))
	} else {
		proxyServerList.ProxyRWMutex.Lock()
		proxyServer, found := proxyServerList.ProxyServerList[proxy.Pid]
		if !found {
			log.Println("proxy with pid " + strconv.Itoa(proxy.Pid) + " not found!")
			proxyServerList.ProxyRWMutex.Unlock()
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"proxy with pid ` + strconv.Itoa(proxy.Pid) + ` not found!"}`))
		} else {
			delete(proxyServerList.ProxyServerList, proxy.Pid)
			proxyServerList.ProxyRWMutex.Unlock()
			proxyServer.Listener.Close()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"proxy with pid ` + strconv.Itoa(proxy.Pid) + ` delete successful!"}`))
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	_, err := auth.JwtAuthRequest(r)
	if err != nil {
		log.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Access Unauthorized!"}`))
		return
	}
	token, err := r.Cookie("kubernetes_token")
	if err != nil {
		log.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal server error!"}`))
		return
	}
	r.ParseForm()
	req, err := http.NewRequest(r.Method, proxyAddr+"?pid="+r.FormValue("pid"), r.Body)
	if err != nil {
		log.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal server error!"}`))
		return
	}
	req.AddCookie(token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal server error!"}`))
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal server error!"}`))
		return
	}
	header := w.Header()
	for k, v := range res.Header {
		header[k] = v
	}
	w.WriteHeader(res.StatusCode)
	w.Write(body)
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
	if r.Method == "GET" {
		queryProxyServer(w, r, claim)
		return
	}
	if r.Method == "POST" {
		createProxyServer(w, r, claim)
		return
	}
	if r.Method == "DELETE" {
		deleteProxyServer(w, r, claim)
		return
	}
	log.Println("only method GET, POST and DELETE are supported")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"message":"only method GET, POST and DELETE are supported"}`))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"health"}`))
}

func main() {
	var (
		proxy     Proxy
		pid       int
		uid       int
		proxyName string
		destIP    string
		destPort  int
		protocol  string
		suffix    string
		wsSuffix  string
	)
	proxyRWMutex = new(sync.RWMutex)
	proxyServerMap = make(map[int]*ProxyServerList)
	tp = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	config := new(conf.Config)
	err := config.InitConfig("proxy/proxy.ini")
	if err != nil {
		log.Fatalln(err)
	}
	logFileName := config.Get("logFile")
	if logFileName == "" {
		log.Fatalln("logFile must be set")
	}
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	ip = config.Get("ip")
	if ip == "" {
		log.Fatalln("ip must be set")
	}
	port := config.Get("port")
	if port == "" {
		log.Fatalln("port must be set")
	}
	upstream = config.Get("upstream")
	if upstream != "" {
		req, err := http.NewRequest("GET", upstream+"/health", nil)
		if err != nil {
			log.Fatalln(err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatalln(err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalln("Please make sure upstream proxy server " + upstream + " is running and state is health!")
		}
	}
	proxyAddr = config.Get("proxyAddr")
	if proxyAddr == "" {
		log.Fatalln("proxyAddr must be set")
	}
	privateKeyPath := config.Get("privateKey")
	if privateKeyPath == "" {
		log.Fatalln("privatekey must be set")
	}
	publicKeyPath := config.Get("publicKey")
	if publicKeyPath == "" {
		log.Fatalln("publicKey must be set")
	}
	err = auth.JwtInit(privateKeyPath, publicKeyPath)
	if err != nil {
		log.Fatalln(err)
	}
	certFile = config.Get("certFile")
	if certFile == "" {
		log.Fatalln("certFile must be set")
	}
	keyFile = config.Get("keyFile")
	if keyFile == "" {
		log.Fatalln("keyFile must be set")
	}
	dbStr := config.Get("dbStr")
	if dbStr == "" {
		log.Fatalln("dbStr must be set")
	}
	err = db.InitMysqlDB(dbStr)
	if err != nil {
		log.Fatalln(err)
	}
	if upstream == "" {
		proxys, err := db.MysqlDB.Query("select pid, uid, proxyName, ip, port, protocol, suffix, wsSuffix from apps, proxys where apps.aid=proxys.aid")
		if err != nil {
			log.Fatalln(err)
		}
		defer proxys.Close()
		for proxys.Next() {
			err = proxys.Scan(&pid, &uid, &proxyName, &destIP, &destPort, &protocol, &suffix, &wsSuffix)
			if err != nil {
				log.Fatalln(err)
			}
			serveMux := http.NewServeMux()
			proxyServer := new(ProxyServer)
			proxyServer.Uid = uid
			proxyServer.Listener, err = net.Listen("tcp", ip+":0")
			if err != nil {
				log.Fatalln(err)
			}
			proxyServer.ProxyName = proxyName
			proxyServer.Protocol = protocol
			proxyServer.Suffix = suffix
			proxyServer.WSSuffix = wsSuffix
			proxyServer.Port = proxyServer.Listener.Addr().(*net.TCPAddr).Port
			if protocol == "http" {
				remote, err := url.Parse("http://" + destIP + ":" + strconv.Itoa(destPort))
				if err != nil {
					log.Fatalln(err)
				}
				proxyServer.HttpProxy = httputil.NewSingleHostReverseProxy(remote)
				serveMux.HandleFunc("/", proxyServer.HttpProxyHandler)
			} else {
				remote, err := url.Parse("https://" + destIP + ":" + strconv.Itoa(destPort))
				if err != nil {
					log.Fatalln(err)
				}
				proxyServer.HttpsProxy = httputil.NewSingleHostReverseProxy(remote)
				proxyServer.HttpsProxy.Transport = tp
				serveMux.HandleFunc("/", proxyServer.HttpsProxyHandler)
			}
			if wsSuffix != "" {
				remote, err := url.Parse("ws://" + destIP + ":" + strconv.Itoa(destPort) + "/" + suffix)
				if err != nil {
					log.Fatalln(err)
				}
				proxyServer.WebSocketProxy = websocketproxy.NewProxy(remote)
				proxyServer.WebSocketProxy.Director = func(req *http.Request, requsetHeader http.Header) {
					requsetHeader.Add("Host", req.Header.Get("Host"))
				}
				serveMux.HandleFunc("/"+wsSuffix, proxyServer.WebSocketProxyHandler)
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
			server := &http.Server{Handler: serveMux}
			go func() {
				proxyServer.errChan = make(chan error, 1)
				if protocol == "http" {
					proxyServer.errChan <- server.Serve(tcpKeepAliveListener{proxyServer.Listener.(*net.TCPListener)})
				} else {
					proxyServer.errChan <- server.ServeTLS(tcpKeepAliveListener{proxyServer.Listener.(*net.TCPListener)}, certFile, keyFile)
				}
			}()
		}
	} else {
		proxys, err := db.MysqlDB.Query("select pid, uid from apps, proxys where apps.aid=proxys.aid")
		if err != nil {
			log.Fatalln(err)
		}
		defer proxys.Close()
		for proxys.Next() {
			err = proxys.Scan(&pid, &uid)
			if err != nil {
				log.Fatalln(err)
			}
			token, err := auth.JwtCreateToken(uid, "", 0)
			if err != nil {
				log.Fatalln(err)
			}
			req, err := http.NewRequest("GET", upstream+"/proxy?pid="+strconv.Itoa(pid), nil)
			if err != nil {
				log.Fatalln(err)
			}
			req.AddCookie(&http.Cookie{Name: "kubernetes_token", Value: token, Expires: time.Now().Add(time.Hour * 24)})
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Fatalln(err)
			}
			data, _ := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if res.StatusCode != http.StatusOK {
				log.Fatalln("failed to query proxy info!")
			}
			err = json.Unmarshal(data, &proxy)
			if err != nil {
				log.Fatalln(err)
			}
			proxyName, destIP, destPort, protocol, suffix, wsSuffix = proxy.ProxyName, proxy.IP, proxy.Port, proxy.Protocol, proxy.Suffix, proxy.WSSuffix
			serveMux := http.NewServeMux()
			proxyServer := new(ProxyServer)
			proxyServer.Uid = uid
			proxyServer.Listener, err = net.Listen("tcp", ip+":0")
			if err != nil {
				log.Fatalln(err)
			}
			proxyServer.ProxyName = proxyName
			proxyServer.Protocol = protocol
			proxyServer.Suffix = suffix
			proxyServer.WSSuffix = wsSuffix
			proxyServer.Port = proxyServer.Listener.Addr().(*net.TCPAddr).Port
			if protocol == "http" {
				remote, err := url.Parse("http://" + destIP + ":" + strconv.Itoa(destPort))
				if err != nil {
					log.Fatalln(err)
				}
				proxyServer.HttpProxy = httputil.NewSingleHostReverseProxy(remote)
				serveMux.HandleFunc("/", proxyServer.HttpProxyHandler)
			} else {
				remote, err := url.Parse("https://" + destIP + ":" + strconv.Itoa(destPort))
				if err != nil {
					log.Fatalln(err)
				}
				proxyServer.HttpsProxy = httputil.NewSingleHostReverseProxy(remote)
				proxyServer.HttpsProxy.Transport = tp
				serveMux.HandleFunc("/", proxyServer.HttpsProxyHandler)
			}
			if wsSuffix != "" {
				remote, err := url.Parse("ws://" + destIP + ":" + strconv.Itoa(destPort) + "/" + suffix)
				if err != nil {
					log.Fatalln(err)
				}
				proxyServer.WebSocketProxy = websocketproxy.NewProxy(remote)
				proxyServer.WebSocketProxy.Director = func(req *http.Request, requsetHeader http.Header) {
					requsetHeader.Add("Host", req.Header.Get("Host"))
				}
				serveMux.HandleFunc("/"+wsSuffix, proxyServer.WebSocketProxyHandler)
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
			server := &http.Server{Handler: serveMux}
			go func() {
				proxyServer.errChan = make(chan error, 1)
				if protocol == "http" {
					proxyServer.errChan <- server.Serve(tcpKeepAliveListener{proxyServer.Listener.(*net.TCPListener)})
				} else {
					proxyServer.errChan <- server.ServeTLS(tcpKeepAliveListener{proxyServer.Listener.(*net.TCPListener)}, certFile, keyFile)
				}
			}()
		}
	}
	httpMux := http.NewServeMux()
	if upstream == "" {
		httpMux.HandleFunc("/", handler)
	}
	httpMux.HandleFunc("/proxy", proxyHandler)
	httpMux.HandleFunc("/health", healthHandler)
	server := &http.Server{Addr: ip + ":" + port, Handler: httpMux}
	log.Println("Starting Proxy Server ...")
	log.Println("Listening on address " + ip + ":" + port)
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ListenAndServe()
		log.Println("shutting down proxy server ...")
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
