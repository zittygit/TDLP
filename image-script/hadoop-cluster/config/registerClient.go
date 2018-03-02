package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"net"
)
var (
	mutex sync.Mutex
)
type Host struct {
	IP string `json:"ip"`
	HostName string `json:hostname`
}
func registeHostName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"message":"only method POST is supported"}`))
		return
	}
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var host Host
	err := json.Unmarshal(data, &host)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error"}`))
		return
	}
	if host.IP == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"ip is required"}`))
		return
	}
	if host.HostName == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"hostname is required"}`))
		return
	}
	mutex.Lock()
	hosts, err := os.OpenFile("/etc/hosts", os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("failed to open parameter server file")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to update parameter server config"}`))
		mutex.Unlock()
		return
	}
	w.WriteHeader(http.StatusOK)
	hosts.WriteString("\n"+host.IP+"   "+host.HostName)
	hosts.Close()
	mutex.Unlock()

}

func sendMsg(serverAddr string,ip string,hostname string){
	for {
		req, err := http.NewRequest("POST", serverAddr, strings.NewReader(`{"ip":"`+ip+`"`+`,"hostname":"`+hostname+`"}`))
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second * 1)
			continue
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second * 1)
			continue
		}
		if res.StatusCode == 200 {
			break
		} else {
			time.Sleep(time.Second * 1)
		}
	}
	fmt.Println(hostname+" register done")
}
func main() {
	var (
		serverAddr string
		IP         string
		hostname   string
	)
	if len(os.Args) != 2 {
		fmt.Println("usage: " + os.Args[0] + " serverIP")
		return
	}
	serverAddr ="http://" +  os.Args[1]+":3333/registerHostName"
	hosts, err := os.Hostname()
	if err != nil {
		fmt.Println("failed to get host address")
		return
	}
	hostname = hosts
	addrs, err := net.LookupHost(hostname)
	if err != nil {
		fmt.Println("failed to get host address")
		return
	}
	for _, addr := range addrs {
		IP = addr
	}
	go sendMsg(serverAddr,IP,hostname)
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/registerHostName", registeHostName)
	server := &http.Server{Addr: ":3333", Handler: serveMux}
	server.ListenAndServe()

}
