package main
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"strings"
)

var (
	mutex sync.Mutex
	slaves []Slaves
)

type Host struct {
	IP string `json:"ip"`
	HostName string `json:hostname`
}

type Slaves struct {
	IP string
	Data string
}

func slaveSync(ip string,data string){
	for i:=0;i<len(slaves);i++ {
		if(slaves[i].IP!=ip){
			//将已有的IP同步到当前slave
			req, err := http.NewRequest("POST","http://" + ip+":3333/registerHostName" , strings.NewReader(slaves[i].Data))
			if err != nil {
				fmt.Println(err)
				return
			}
			res, err := http.DefaultClient.Do(req)
			if res.StatusCode != 200 {
				fmt.Println(ip+" sync error")
			}
			//将当前slave IP同步到其他slave
			req, err = http.NewRequest("POST","http://" + slaves[i].IP+":3333/registerHostName" , strings.NewReader(data))
			if err != nil {
				fmt.Println(err)
				return
			}
			res, err = http.DefaultClient.Do(req)
			if res.StatusCode != 200 {
				fmt.Println(slaves[i].IP+" sync error")
			}
		}
	}
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
	hosts.WriteString(host.IP+"   "+host.HostName+"\n")
	hosts.Close()
	slaveSync(host.IP,`{"ip":"`+host.IP+`"`+`,"hostname":"`+host.HostName+`"}`)
	slave := Slaves{IP:host.IP,Data:`{"ip":"`+host.IP+`"`+`,"hostname":"`+host.HostName+`"}`}
	slaves = append(slaves,slave)
	mutex.Unlock()
}



func main() {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/registerHostName", registeHostName)
	server := &http.Server{Addr: ":3333", Handler: serveMux}
	server.ListenAndServe()
}
