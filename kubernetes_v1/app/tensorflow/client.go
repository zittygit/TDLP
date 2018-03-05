package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	hostname   string
	job_name   string
	task_index int
)

type Script struct {
	Path string `json:"path"`
}

type Index struct {
	Index int `json:"index"`
}

func runScript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var script Script
	err := json.Unmarshal(data, &script)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error"}`))
		return
	}
	if script.Path == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"script path is required"}`))
		return
	}
	ln, err := net.Listen("tcp", ":2222")
	if err == nil {
		err = ln.Close()
		if err != nil {
			fmt.Println("couldn't close port 2222")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"Internal Server Error!"`))
			return
		}
		go func() {
			cmd := exec.Command("python", script.Path, job_name, strconv.Itoa(task_index))
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
	fmt.Println("running with command: python " + script.Path + " " + job_name + " " + strconv.Itoa(task_index))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"script is running"}`))
}

func main() {
	var (
		serverAddr string
		IP         string
	)
	if len(os.Args) != 3 {
		fmt.Println("usage: " + os.Args[0] + " serverIP job_name")
		return
	}
	serverIP := os.Args[1]
	job_name = os.Args[2]
	if job_name == "ps" {
		serverAddr = "http://" + serverIP + ":3333/registeParameter"
	} else {
		serverAddr = "http://" + serverIP + ":3333/registeWorker"
	}
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
		fmt.Println(IP)
	}
	for {
		req, err := http.NewRequest("POST", serverAddr, strings.NewReader(`{"ip":"`+IP+`"}`))
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second * 5)
			continue
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second * 5)
			continue
		}
		if res.StatusCode == 200 {
			data, _ := ioutil.ReadAll(res.Body)
			res.Body.Close()
			var index Index
			err := json.Unmarshal(data, &index)
			if err != nil {
				fmt.Println(err)
				time.Sleep(time.Second * 5)
			} else {
				task_index = index.Index
				fmt.Println("registe to server " + serverIP + " with index " + strconv.Itoa(task_index))
				break
			}
		} else {
			time.Sleep(time.Second * 5)
		}
	}
	serveMux := http.NewServeMux()
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
