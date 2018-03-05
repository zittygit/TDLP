package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"kubernetes/conf"
	"kubernetes/k8s"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Param struct {
	Cpu    int `json:"cpu"`
	Memory int `json:"memory"`
}

func create() {
	if flag.NArg() != 4 {
		fmt.Println(`{"message": "usage: ` + os.Args[0] + ` --action create uid userName instancename param"}`)
		os.Exit(1)
	}
	uid := flag.Arg(0)
	userName := flag.Arg(1)
	instanceName := flag.Arg(2)
	data := flag.Arg(3)
	match, _ := regexp.MatchString(`^[1-9][0-9]+`, uid)
	if !match {
		fmt.Println(`{"message":"uid format error"}`)
		os.Exit(1)
	}
	match, _ = regexp.MatchString(`^[A-Za-z][-_0-9A-Za-z]+`, userName)
	if !match {
		fmt.Println(`{"message":"userName format error"}`)
		os.Exit(1)
	}
	match, _ = regexp.MatchString(`^[A-Za-z][-_0-9A-Za-z]+`, instanceName)
	if !match {
		fmt.Println(`{"message":"instanceName format error"}`)
		os.Exit(1)
	}
	var param Param
	err := json.Unmarshal([]byte(data), &param)
	if err != nil {
		fmt.Println(`{"message":"param format error"}`)
		os.Exit(1)
	}
	if param.Cpu == 0 || param.Memory == 0 {
		fmt.Println(`{"message":"cpu and memory are required"}`)
		os.Exit(1)
	}
	name := instanceName + "-tensorflow-cpu"
	req, err := http.NewRequest("POST", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers", strings.NewReader(`{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":1,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["/tensorflow.sh"],"image":"nscc/tensorflow:1.3.0-cpu","name":"`+name+`","ports":[{"containerPort":8000},{"containerPort":8080},{"containerPort":8888}],"resources":{"requests":{"cpu":"`+strconv.Itoa(param.Cpu)+`m","memory":"`+strconv.Itoa(param.Memory)+`Mi"}}}],"securityContext":{"runAsUser":`+uid+`}}}}}`))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if res.StatusCode != 201 {
		buf, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		fmt.Println(string(buf))
		os.Exit(1)
	}
	req, err = http.NewRequest("POST", k8s.K8sApiServer+"/namespaces/"+userName+"/services", strings.NewReader(`{"apiVersion":"v1","kind":"Service","metadata":{"name":"`+name+`"},"spec":{"ports":[{"name":"terminal","port":8000,"targetPort":8000},{"name":"jupyter","port":8080,"targetPort":8080},{"name":"tensorboard","port":8888,"targetPort":8888}],"selector":{"component":"`+name+`"}}}`))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	res, err = client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	buf, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode == 201 {
		ip := gjson.Get(string(buf), "spec.clusterIP").String()
		fmt.Println(`{"services":[{"proxyname":"tensorflow-ssh","httpurl":"http://` + ip + `:8000","websocketurl":"ws://` + ip + `:8000"},{"proxyname":"jupyter-web","httpurl":"http://` + ip + `:8080"},{"proxyname":"tensorboard","httpurl":"http://` + ip + `:8888"}]}`)
	} else {
		fmt.Println(string(buf))
		os.Exit(1)
	}
}

func update() {
	if flag.NArg() != 4 {
		fmt.Println(`{"message": "usage: ` + os.Args[0] + ` --action update uid userName instancename param"}`)
		os.Exit(1)
	}
	uid := flag.Arg(0)
	userName := flag.Arg(1)
	instanceName := flag.Arg(2)
	data := flag.Arg(3)
	match, _ := regexp.MatchString(`^[1-9][0-9]+`, uid)
	if !match {
		fmt.Println(`{"message":"uid format error"}`)
		os.Exit(1)
	}
	match, _ = regexp.MatchString(`^[A-Za-z][-_0-9A-Za-z]+`, userName)
	if !match {
		fmt.Println(`{"message":"userName format error"}`)
		os.Exit(1)
	}
	match, _ = regexp.MatchString(`^[A-Za-z][-_0-9A-Za-z]+`, instanceName)
	if !match {
		fmt.Println(`{"message":"instanceName format error"}`)
		os.Exit(1)
	}
	var param Param
	err := json.Unmarshal([]byte(data), &param)
	if err != nil {
		fmt.Println(`{"message":"param format error"}`)
		os.Exit(1)
	}
	if param.Cpu == 0 || param.Memory == 0 {
		fmt.Println(`{"message":"cpu and memory are required"}`)
		os.Exit(1)
	}
	name := instanceName + "-tensorflow-cpu"
	req, err := http.NewRequest("PUT", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, strings.NewReader(`{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":0,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["/tensorflow.sh"],"image":"nscc/tensorflow:1.3.0-cpu","name":"`+name+`","ports":[{"containerPort":8000},{"containerPort":8080},{"containerPort":8888}]}]}}}}`))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if res.StatusCode != 200 {
		data, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		fmt.Println(string(data))
		os.Exit(1)
	}
	req, err = http.NewRequest("PUT", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, strings.NewReader(`{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":1,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["/tensorflow.sh"],"image":"nscc/tensorflow:1.3.0-cpu","name":"`+name+`","ports":[{"containerPort":8000},{"containerPort":8080},{"containerPort":8888}],"resources":{"requests":{"cpu":"`+strconv.Itoa(param.Cpu)+`m","memory":"`+strconv.Itoa(param.Memory)+`Mi"}}}],"securityContext":{"runAsUser":`+uid+`}}}}}`))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	res, err = client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if res.StatusCode == 200 {
		fmt.Println(`{"message":"updated successful"}`)
	} else {
		data, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		fmt.Println(string(data))
		os.Exit(1)
	}
}

func delete() {
	if flag.NArg() != 2 {
		fmt.Println(`{"message": "usage: ` + os.Args[0] + ` --action delete userName instancename"}`)
		os.Exit(1)
	}
	userName := flag.Arg(0)
	instanceName := flag.Arg(1)
	match, _ := regexp.MatchString(`^[A-Za-z][-_0-9A-Za-z]+`, userName)
	if !match {
		fmt.Println(`{"message":"userName format error"}`)
		os.Exit(1)
	}
	match, _ = regexp.MatchString(`^[A-Za-z][-_0-9A-Za-z]+`, instanceName)
	if !match {
		fmt.Println(`{"message":"instanceName format error"}`)
		os.Exit(1)
	}
	name := instanceName + "-tensorflow-cpu"
	req, err := http.NewRequest("PUT", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, strings.NewReader(`{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":0,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["/tensorflow.sh"],"image":"nscc/tensorflow:1.3.0-cpu","name":"`+name+`","ports":[{"containerPort":8000},{"containerPort":8080},{"containerPort":8888}]}]}}}}`))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if res.StatusCode != 200 {
		data, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		fmt.Println(string(data))
		os.Exit(1)
	}
	req, err = http.NewRequest("DELETE", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	res, err = client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if res.StatusCode != 200 {
		data, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		fmt.Println(string(data))
		os.Exit(1)
	}
	req, err = http.NewRequest("DELETE", k8s.K8sApiServer+"/namespaces/"+userName+"/services/"+name, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	res, err = client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode == 200 {
		fmt.Println(`{"message":"deleted successful"}`)
	} else {
		fmt.Println(string(data))
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`{"message": "usage: ` + os.Args[0] + ` --action create|update|delete"}`)
}

func main() {
	configFile := "k8s/k8s.ini"
	action := flag.String("action", "create", "action for manage tensorflow")
	flag.Parse()
	myConfig := new(conf.Config)
	myConfig.InitConfig(&configFile)
	k8s.BearerToken = myConfig.Read("bearertoken")
	if k8s.BearerToken == "" {
		k8s.BearerToken = "Bearer 6B1GbqhcjqGYPAAy285otYhUUV4z4kiu"
	}
	k8s.K8sApiServer = myConfig.Read("k8sapiserver")
	if k8s.K8sApiServer == "" {
		k8s.K8sApiServer = "https://10.127.48.18:6443/api/v1"
	}
	switch {
	case *action == "create":
		create()
	case *action == "update":
		update()
	case *action == "delete":
		delete()
	default:
		usage()
	}
}
