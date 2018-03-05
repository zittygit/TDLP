package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
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
	Cpu              int `json:"cpu"`
	Memory           int `json:"memory"`
	ParameterServers int `json:"parameterservers"`
	WorkerServers    int `json:"workerservers"`
}

func createClient(uid string, userName string, instanceName string, cpu string, memory string, parameterServers string, workerServers string) error {
	name := instanceName + "-tensorflow-client"
	volume := instanceName + "-tensorflow"
	req, err := http.NewRequest("POST", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers", strings.NewReader(`{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":1,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["/tensorflow-client.sh","`+parameterServers+`","`+workerServers+`"],"image":"nscc/tensorflow:1.3.0-cpu-cluster","name":"`+name+`","ports":[{"containerPort":3333},{"containerPort":8000},{"containerPort":8080},{"containerPort":8888}],"resources":{"requests":{"cpu":"`+cpu+`m","memory":"`+memory+`Mi"}},"volumeMounts":[{"mountPath":"/tensorflow","name":"`+volume+`"}]}],"volumes":[{"name":"`+volume+`","hostPath":{"path":"/root/tensorflow"}}],"securityContext":{"runAsUser":`+uid+`}}}}}`))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == 201 {
		return nil
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	return errors.New(string(data))
}

func updateClient(uid string, userName string, instanceName string, cpu string, memory string, parameterServers string, workerServers string) error {
	name := instanceName + "-tensorflow-client"
	volume := instanceName + "-tensorflow"
	req, err := http.NewRequest("PUT", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, strings.NewReader(`{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":0,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["/tensorflow-client.sh"],"image":"nscc/tensorflow:1.3.0-cpu-cluster","name":"`+name+`","ports":[{"containerPort":3333},{"containerPort":8000},{"containerPort":8080},{"containerPort":8888}]}]}}}}`))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		data, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		return errors.New(string(data))
	}
	req, err = http.NewRequest("PUT", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, strings.NewReader(`{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":1,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["/tensorflow-client.sh","`+parameterServers+`","`+workerServers+`"],"image":"nscc/tensorflow:1.3.0-cpu-cluster","name":"`+name+`","ports":[{"containerPort":3333},{"containerPort":8000},{"containerPort":8080},{"containerPort":8888}],"resources":{"requests":{"cpu":"`+cpu+`m","memory":"`+memory+`Mi"}},"volumeMounts":[{"mountPath":"/tensorflow","name":"`+volume+`"}]}],"volumes":[{"name":"`+volume+`","hostPath":{"path":"/root/tensorflow"}}],"securityContext":{"runAsUser":`+uid+`}}}}}`))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	res, err = client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == 200 {
		return nil
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	return errors.New(string(data))
}

func deleteClient(userName string, instanceName string) error {
	name := instanceName + "-tensorflow-client"
	req, err := http.NewRequest("PUT", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, strings.NewReader(`{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":0,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["/tensorflow-client.sh"],"image":"nscc/tensorflow:1.3.0-cpu-cluster","name":"`+name+`","ports":[{"containerPort":3333},{"containerPort":8000},{"containerPort":8080},{"containerPort":8888}]}]}}}}`))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		data, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		return errors.New(string(data))
	}
	req, err = http.NewRequest("DELETE", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	res, err = client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == 200 {
		return nil
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	return errors.New(string(data))
}

func createClientService(userName string, instanceName string) (string, error) {
	name := instanceName + "-tensorflow-client"
	req, err := http.NewRequest("POST", k8s.K8sApiServer+"/namespaces/"+userName+"/services", strings.NewReader(`{"apiVersion":"v1","kind":"Service","metadata":{"name":"`+name+`"},"spec":{"ports":[{"name":"server","port":3333,"targetPort":3333},{"name":"ssh","port":8000,"targetPort":8000},{"name":"jupyter","port":8080,"targetPort":8080},{"name":"tensorboard","port":8888,"targetPort":8888}],"selector":{"component":"`+name+`"}}}`))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode == 201 {
		ip := gjson.Get(string(data), "spec.clusterIP").String()
		return ip, nil
	} else {
		return "", errors.New(string(data))
	}
}

func deleteClientService(userName string, instanceName string) error {
	name := instanceName + "-tensorflow-client"
	req, err := http.NewRequest("DELETE", k8s.K8sApiServer+"/namespaces/"+userName+"/services/"+name, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode == 200 {
		return nil
	} else {
		return errors.New(string(data))
	}
}

func createParameterServer(uid string, userName string, instanceName string, cpu string, memory string, clientIP string, parameterServers string) error {
	name := instanceName + "-tensorflow-parameter"
	volume := instanceName + "-tensorflow"
	req, err := http.NewRequest("POST", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers", strings.NewReader(`{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":`+parameterServers+`,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["/tensorflow-parameter.sh","`+clientIP+`"],"image":"nscc/tensorflow:1.3.0-cpu-cluster","name":"`+name+`","ports":[{"containerPort":2222},{"containerPort":3333}],"resources":{"requests":{"cpu":"`+cpu+`m","memory":"`+memory+`Mi"}},"volumeMounts":[{"mountPath":"/tensorflow","name":"`+volume+`"}]}],"volumes":[{"name":"`+volume+`","hostPath":{"path":"/root/tensorflow"}}],"securityContext":{"runAsUser":`+uid+`}}}}}`))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == 201 {
		return nil
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	return errors.New(string(data))
}

func updateParameterServer(uid string, userName string, instanceName string, cpu string, memory string, clientIP string, parameterServers string) error {
	name := instanceName + "-tensorflow-parameter"
	volume := instanceName + "-tensorflow"
	req, err := http.NewRequest("PUT", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, strings.NewReader(`{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":0,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["/tensorflow-parameter.sh"],"image":"nscc/tensorflow:1.3.0-cpu-cluster","name":"`+name+`","ports":[{"containerPort":2222},{"containerPort":3333}]}]}}}}`))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		data, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		return errors.New(string(data))
	}
	req, err = http.NewRequest("PUT", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, strings.NewReader(`{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":`+parameterServers+`,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["/tensorflow-parameter.sh","`+clientIP+`"],"image":"nscc/tensorflow:1.3.0-cpu-cluster","name":"`+name+`","ports":[{"containerPort":2222},{"containerPort":3333}],"resources":{"requests":{"cpu":"`+cpu+`m","memory":"`+memory+`Mi"}},"volumeMounts":[{"mountPath":"/tensorflow","name":"`+volume+`"}]}],"volumes":[{"name":"`+volume+`","hostPath":{"path":"/root/tensorflow"}}],"securityContext":{"runAsUser":`+uid+`}}}}}`))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	res, err = client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == 200 {
		return nil
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	return errors.New(string(data))
}

func deleteParameterServer(userName string, instanceName string) error {
	name := instanceName + "-tensorflow-parameter"
	req, err := http.NewRequest("PUT", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, strings.NewReader(`{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":0,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["/tensorflow-parameter.sh"],"image":"nscc/tensorflow:1.3.0-cpu-cluster","name":"`+name+`","ports":[{"containerPort":2222},{"containerPort":3333}]}]}}}}`))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		data, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		return errors.New(string(data))
	}
	req, err = http.NewRequest("DELETE", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	res, err = client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == 200 {
		return nil
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	return errors.New(string(data))
}

func createWorkerServer(uid string, userName string, instanceName string, cpu string, memory string, clientIP string, workerServers string) error {
	name := instanceName + "-tensorflow-worker"
	volume := instanceName + "-tensorflow"
	req, err := http.NewRequest("POST", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers", strings.NewReader(`{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":`+workerServers+`,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["/tensorflow-worker.sh","`+clientIP+`"],"image":"nscc/tensorflow:1.3.0-cpu-cluster","name":"`+name+`","ports":[{"containerPort":2222},{"containerPort":3333}],"resources":{"requests":{"cpu":"`+cpu+`m","memory":"`+memory+`Mi"}},"volumeMounts":[{"mountPath":"/tensorflow","name":"`+volume+`"}]}],"volumes":[{"name":"`+volume+`","hostPath":{"path":"/root/tensorflow"}}],"securityContext":{"runAsUser":`+uid+`}}}}}`))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == 201 {
		return nil
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	return errors.New(string(data))
}

func updateWorkerServer(uid string, userName string, instanceName string, cpu string, memory string, clientIP string, workerServers string) error {
	name := instanceName + "-tensorflow-worker"
	volume := instanceName + "-tensorflow"
	req, err := http.NewRequest("PUT", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, strings.NewReader(`{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":0,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["/tensorflow-worker.sh"],"image":"nscc/tensorflow:1.3.0-cpu-cluster","name":"`+name+`","ports":[{"containerPort":2222},{"containerPort":3333}]}]}}}}`))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		data, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		return errors.New(string(data))
	}
	req, err = http.NewRequest("PUT", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, strings.NewReader(`{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":`+workerServers+`,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["/tensorflow-worker.sh","`+clientIP+`"],"image":"nscc/tensorflow:1.3.0-cpu-cluster","name":"`+name+`","ports":[{"containerPort":2222},{"containerPort":3333}],"resources":{"requests":{"cpu":"`+cpu+`m","memory":"`+memory+`Mi"}},"volumeMounts":[{"mountPath":"/tensorflow","name":"`+volume+`"}]}],"volumes":[{"name":"`+volume+`","hostPath":{"path":"/root/tensorflow"}}],"securityContext":{"runAsUser":`+uid+`}}}}}`))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	res, err = client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == 200 {
		return nil
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	return errors.New(string(data))
}

func deleteWorkerServer(userName string, instanceName string) error {
	name := instanceName + "-tensorflow-worker"
	req, err := http.NewRequest("PUT", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, strings.NewReader(`{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":0,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["/tensorflow-worker.sh"],"image":"nscc/tensorflow:1.3.0-cpu-cluster","name":"`+name+`","ports":[{"containerPort":2222},{"containerPort":3333}]}]}}}}`))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		data, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		return errors.New(string(data))
	}
	req, err = http.NewRequest("DELETE", k8s.K8sApiServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	res, err = client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == 200 {
		return nil
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	return errors.New(string(data))
}

func createCluster(uid string, userName string, instanceName string, cpu string, memory string, parameterServers string, workerServers string) (string, error) {
	err := createClient(uid, userName, instanceName, cpu, memory, parameterServers, workerServers)
	if err != nil {
		return "", err
	}
	ip, err := createClientService(userName, instanceName)
	if err != nil {
		deleteClient(userName, instanceName)
		return "", err
	}
	err = createParameterServer(uid, userName, instanceName, cpu, memory, ip, parameterServers)
	if err != nil {
		deleteClient(userName, instanceName)
		deleteClientService(userName, instanceName)
		return "", err
	}
	err = createWorkerServer(uid, userName, instanceName, cpu, memory, ip, workerServers)
	if err != nil {
		deleteClient(userName, instanceName)
		deleteClientService(userName, instanceName)
		deleteParameterServer(userName, instanceName)
		return "", err
	} else {
		return ip, err
	}
}

func updateCluster(uid string, userName string, instanceName string, cpu string, memory string, parameterServers string, workerServers string) error {
	name := instanceName + "-tensorflow-client"
	req, err := http.NewRequest("GET", k8s.K8sApiServer+"/namespaces/"+userName+"/services/"+name, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", k8s.BearerToken)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != 200 {
		return errors.New(string(data))
	}
	clientIP := gjson.Get(string(data), "spec.clusterIP").String()
	err = updateClient(uid, userName, instanceName, cpu, memory, parameterServers, workerServers)
	if err != nil {
		return err
	}
	err = updateParameterServer(uid, userName, instanceName, cpu, memory, clientIP, parameterServers)
	if err != nil {
		return err
	}
	return updateWorkerServer(uid, userName, instanceName, cpu, memory, clientIP, workerServers)
}

func deleteCluster(userName string, instanceName string) error {
	err := deleteWorkerServer(userName, instanceName)
	if err != nil {
		return err
	}
	err = deleteParameterServer(userName, instanceName)
	if err != nil {
		return err
	}
	err = deleteClient(userName, instanceName)
	if err != nil {
		return err
	}
	return deleteClientService(userName, instanceName)
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
	if param.Cpu == 0 || param.Memory == 0 || param.ParameterServers == 0 || param.WorkerServers == 0 {
		fmt.Println(`{"message":"cpu, memory, parameterServers and workerServers are required"}`)
		os.Exit(1)
	}
	ip, err := createCluster(uid, userName, instanceName, strconv.Itoa(param.Cpu), strconv.Itoa(param.Memory), strconv.Itoa(param.ParameterServers), strconv.Itoa(param.WorkerServers))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Println(`{"services":[{"proxyname":"tensorflow-ssh","httpurl":"http://` + ip + `:8000","websocketurl":"ws://` + ip + `:8000"},{"proxyname":"jupyter-web","httpurl":"http://` + ip + `:8080"},{"proxyname":"tensorboard","httpurl":"http://` + ip + `:8888"}]}`)
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
	if param.Cpu == 0 || param.Memory == 0 || param.ParameterServers == 0 || param.WorkerServers == 0 {
		fmt.Println(`{"message":"cpu, memory, parameterServers and workerServers are required"}`)
		os.Exit(1)
	}
	err = updateCluster(uid, userName, instanceName, strconv.Itoa(param.Cpu), strconv.Itoa(param.Memory), strconv.Itoa(param.ParameterServers), strconv.Itoa(param.WorkerServers))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Println(`{"message":"updated successful"}`)
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
	err := deleteCluster(userName, instanceName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Println(`{"message":"deleted successful"}`)
	}
}

func usage() {
	fmt.Println(`{"message": "usage: ` + os.Args[0] + ` --action create|update|delete"}`)
}

func main() {
	configFile := "k8s/k8s.ini"
	action := flag.String("action", "create", "action for manage spark cluster")
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
