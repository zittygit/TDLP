package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var (
	hostname             string
	submitting           bool
	parameterServerCount int
	workerServerCount    int
	parameterServers     []string
	workerServers        []string
	parameterMutex       *sync.Mutex
	workerMutex          *sync.Mutex
	configMutex          *sync.Mutex
)

type Host struct {
	IP string `json:"ip"`
}

type Script struct {
	Path string `json:"path"`
}

func runScript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	configMutex.Lock()
	if submitting {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"message":"previous submitting not finished yet"}`))
		configMutex.Unlock()
		return
	} else {
		submitting = true
		configMutex.Unlock()
	}
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var script Script
	err := json.Unmarshal(data, &script)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error"}`))
		configMutex.Lock()
		submitting = false
		configMutex.Unlock()
		return
	}
	if script.Path == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"script path is required"}`))
		configMutex.Lock()
		submitting = false
		configMutex.Unlock()
		return
	}
	_, err = os.Stat(script.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"script path is not correct"}`))
		configMutex.Lock()
		submitting = false
		configMutex.Unlock()
		return
	}
	configFile, err := os.OpenFile("/tensorflow/config/config.py", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("failed to open config file")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to update config"}`))
		configMutex.Lock()
		submitting = false
		configMutex.Unlock()
		return
	}
	configFile.Truncate(0)
	configFile.WriteString("import tensorflow as tf\n\n")
	configFile.WriteString("def init_cluster(job_name, task_index):\n")
	parameterMutex.Lock()
	workerMutex.Lock()
	ps := "\tps_hosts=["
	if parameterServerCount > 0 {
		configFile.WriteString("\tps0 = \"" + parameterServers[0] + ":2222\"\n")
		ps += "ps0"
	}
	for i := 1; i < parameterServerCount; i++ {
		configFile.WriteString("\tps" + strconv.Itoa(i) + " = \"" + parameterServers[i] + ":2222\"\n")
		ps += ", ps" + strconv.Itoa(i)
	}
	ps += "]\n"
	configFile.WriteString(ps)
	worker := "\tworker_hosts=["
	if workerServerCount > 0 {
		configFile.WriteString("\tworker0 = \"" + workerServers[0] + ":2222\"\n")
		worker += "worker0"
	}
	for i := 1; i < workerServerCount; i++ {
		configFile.WriteString("\tworker" + strconv.Itoa(i) + " = \"" + workerServers[i] + ":2222\"\n")
		worker += ", worker" + strconv.Itoa(i)
	}
	worker += "]\n"
	configFile.WriteString(worker)
	configFile.WriteString("\tcluster = tf.train.ClusterSpec({\"ps\": ps_hosts, \"worker\": worker_hosts})\n")
	configFile.WriteString("\tserver = tf.train.Server(cluster, job_name, task_index)\n")
	configFile.WriteString("\tif job_name == \"ps\":\n")
	configFile.WriteString("\t\tserver.join()\n")
	configFile.WriteString("\telse:\n")
	configFile.WriteString("\t\treturn cluster, server\n")
	for i := 0; i < parameterServerCount; i++ {
		req, err := http.NewRequest("POST", "http://"+parameterServers[i]+":3333/run", strings.NewReader(`{"path":"`+script.Path+`"}`))
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"Internal Server Error!"`))
			parameterMutex.Unlock()
			workerMutex.Unlock()
			configMutex.Lock()
			submitting = false
			configMutex.Unlock()
			return
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil || res.StatusCode != 200 {
			if err != nil {
				fmt.Println(err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"Internal Server Error!"`))
			parameterMutex.Unlock()
			workerMutex.Unlock()
			configMutex.Lock()
			submitting = false
			configMutex.Unlock()
			return
		}
		fmt.Println("send script to http://" + parameterServers[i] + ":3333/run")
	}
	for i := 1; i < workerServerCount; i++ {
		req, err := http.NewRequest("POST", "http://"+workerServers[i]+":3333/run", strings.NewReader(`{"path":"`+script.Path+`"}`))
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"Internal Server Error!"`))
			parameterMutex.Unlock()
			workerMutex.Unlock()
			configMutex.Lock()
			submitting = false
			configMutex.Unlock()
			return
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil || res.StatusCode != 200 {
			if err != nil {
				fmt.Println(err)
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"Internal Server Error!"`))
			parameterMutex.Unlock()
			workerMutex.Unlock()
			configMutex.Lock()
			submitting = false
			configMutex.Unlock()
			return
		}
		fmt.Println("send script to http://" + workerServers[i] + ":3333/run")
	}
	ln, err := net.Listen("tcp", ":2222")
	if err == nil {
		err = ln.Close()
		if err != nil {
			fmt.Println("couldn't close port 2222")
			parameterMutex.Unlock()
			workerMutex.Unlock()
			configMutex.Lock()
			submitting = false
			configMutex.Unlock()
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"Internal Server Error!"`))
			return
		}
		go func() {
			cmd := exec.Command("python", script.Path, "worker", "0")
			file, err := os.OpenFile(script.Path+"_"+hostname+".log", os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				fmt.Println(err)
				return
			}
			file.Truncate(0)
			defer file.Close()
			cmd.Stdout = file
			cmd.Stderr = file
			err = cmd.Run()
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
	fmt.Println("running with command: python " + script.Path + " worker 0")
	parameterMutex.Unlock()
	workerMutex.Unlock()
	configMutex.Lock()
	submitting = false
	configMutex.Unlock()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"job is submitted"}`))
}

func registeParameterServer(w http.ResponseWriter, r *http.Request) {
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
	index := -1
	parameterMutex.Lock()
	for i := 0; i < parameterServerCount; i++ {
		req, err := http.NewRequest("CHECK", "http://"+parameterServers[i]+":3333", nil)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"Internal Server Error!"`))
			parameterMutex.Unlock()
			return
		}
		_, err = http.DefaultClient.Do(req)
		if err != nil {
			parameterServers[i] = host.IP
			index = i
			fmt.Println("registe parameter server " + host.IP + " whith index " + strconv.Itoa(index))
			break
		}
	}
	if index == -1 {
		index = parameterServerCount
		parameterServerCount += 1
		parameterServers[index] = host.IP
		fmt.Println("registe parameter server " + host.IP + " whith index " + strconv.Itoa(index))
	}
	parameterServerFile, err := os.OpenFile("/tensorflow/config/parameter_servers", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("failed to open parameter server file")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to update parameter server config"}`))
		parameterMutex.Unlock()
		return
	}
	parameterServerFile.Truncate(0)
	for i := 0; i < parameterServerCount; i++ {
		parameterServerFile.WriteString(parameterServers[i] + "\n")
	}
	parameterServerFile.Close()
	parameterMutex.Unlock()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"index":` + strconv.Itoa(index) + `}`))
}

func registeWorkerServer(w http.ResponseWriter, r *http.Request) {
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
	index := -1
	workerMutex.Lock()
	for i := 0; i < workerServerCount; i++ {
		req, err := http.NewRequest("CHECK", "http://"+workerServers[i]+":3333", nil)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"Internal Server Error!"`))
			workerMutex.Unlock()
			return
		}
		_, err = http.DefaultClient.Do(req)
		if err != nil {
			workerServers[i] = host.IP
			index = i
			fmt.Println("registe worker server " + host.IP + " whith index " + strconv.Itoa(index))
			break
		}
	}
	if index == -1 {
		index = workerServerCount
		workerServerCount += 1
		workerServers[index] = host.IP
		fmt.Println("registe worker server " + host.IP + " whith index " + strconv.Itoa(index))
	}
	workerServerFile, err := os.OpenFile("/tensorflow/config/worker_servers", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("failed to open worker server file")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to update worker server config"}`))
		workerMutex.Unlock()
		return
	}
	workerServerFile.Truncate(0)
	for i := 1; i < workerServerCount; i++ {
		workerServerFile.WriteString(workerServers[i] + "\n")
	}
	workerServerFile.Close()
	workerMutex.Unlock()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"index":` + strconv.Itoa(index) + `}`))
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("usage: " + os.Args[0] + " numberParameterServer numberWorkererServer")
		return
	}
	submitting = false
	parameterServerCount = 0
	numberParameterServer, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("numberParameterServer not a number")
		return
	}
	workerServerCount = 1
	numberWorkerServer, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("numberWorkerServer not a number")
		return
	}
	parameterServers = make([]string, numberParameterServer)
	workerServers = make([]string, numberWorkerServer+1)
	parameterMutex = new(sync.Mutex)
	workerMutex = new(sync.Mutex)
	configMutex = new(sync.Mutex)
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
		workerServers[0] = addr
		fmt.Println(workerServers[0])
	}
	scriptFile, err := os.OpenFile("/tensorflow/run.sh", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("failed to open script file")
		return
	}
	scriptFile.Truncate(0)
	scriptFile.WriteString("#! /bin/bash\n\nif [ $# != 1 ]\nthen\n\techo usage $0 script_path\n\texit\nfi\n\ncurl -X POST -d \"{\\\"path\\\":\\\"$1\\\"}\" http://" + workerServers[0] + ":3333/run\n")
	scriptFile.Close()
	parameterServerFile, err := os.OpenFile("/tensorflow/config/parameter_servers", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("failed to open parameter server file")
		return
	}
	reader := bufio.NewReader(parameterServerFile)
	for {
		d, _, err := reader.ReadLine()
		if err != nil {
			if err != io.EOF {
				fmt.Println("failed to read parameter server file")
				parameterServerFile.Close()
				return
			} else {
				break
			}
		} else {
			parameterServers[parameterServerCount] = string(d)
			parameterServerCount += 1
		}
	}
	parameterServerFile.Close()
	workerServerFile, err := os.OpenFile("/tensorflow/config/worker_servers", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("failed to open worker server file")
		return
	}
	reader = bufio.NewReader(workerServerFile)
	for {
		d, _, err := reader.ReadLine()
		if err != nil {
			if err != io.EOF {
				fmt.Println("failed to read worker server file")
				workerServerFile.Close()
				return
			} else {
				break
			}
		} else {
			workerServers[workerServerCount] = string(d)
			workerServerCount += 1
		}
	}
	workerServerFile.Close()
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/registeParameter", registeParameterServer)
	serveMux.HandleFunc("/registeWorker", registeWorkerServer)
	serveMux.HandleFunc("/run", runScript)
	server := &http.Server{Addr: ":3333", Handler: serveMux}
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ListenAndServe()
		fmt.Println("shutting down server ...")
	}()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	ctx := context.Background()
	for {
		select {
		case err := <-errChan:
			if err != nil {
				fmt.Println(err)
				return
			}
		case s := <-signalChan:
			fmt.Println("Captured %v. Exiting ...", s)
			server.Shutdown(ctx)
		}
	}
}
