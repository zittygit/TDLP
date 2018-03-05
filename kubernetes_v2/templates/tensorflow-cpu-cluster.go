package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"kubernetes/k8s"
	"os"
	"regexp"
)

type Param struct {
	CPU              string `json:"cpu"`
	Memory           string `json:"memory"`
	ParameterServers string `json:"parameterservers"`
	WorkerServers    string `json:"workerservers"`
}

func createClientService(userName, instanceName string) (string, error) {
	name := instanceName + "-tensorflow-client"
	return k8s.CreateService(userName, `{"apiVersion":"v1","kind":"Service","metadata":{"name":"`+name+`"},"spec":{"ports":[{"name":"server","port":3333,"targetPort":3333},{"name":"ssh","port":8000,"targetPort":8000},{"name":"jupyter","port":8080,"targetPort":8080},{"name":"tensorboard","port":8888,"targetPort":8888}],"selector":{"component":"`+name+`"}}}`)
}

func deleteClientService(userName, instanceName string) error {
	name := instanceName + "-tensorflow-client"
	return k8s.DeleteService(userName, name)
}

func createClient(userName, instanceName, uid, ip, cpu, memory, parameterServers, workerServers string) error {
	name := instanceName + "-tensorflow-client"
	cwd := "/home/" + userName + "/" + instanceName + "-tensorflow-cpu-cluster"
	return k8s.CreateReplicationController(userName, `{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":1,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["tensorflow-client.sh","`+ip+`","`+parameterServers+`","`+workerServers+`"],"image":"nscc-gz.cn/nscc/tensorflow:1.3.0-cpu-cluster","env":[{"name":"USER","value":"`+userName+`"},{"name":"HOME","value":"/home/`+userName+`"},{"name":"CWD","value":"`+cwd+`"},{"name":"LDAP_SERVER_1","value":"`+k8s.LDAPServer1+`"},{"name":"LDAP_SERVER_2","value":"`+k8s.LDAPServer2+`"},{"name":"TERM","value":"xterm"}],"name":"`+name+`","resources":{"requests":{"cpu":"`+cpu+`m","memory":"`+memory+`Mi"},"limits":{"cpu":"`+cpu+`m","memory":"`+memory+`Mi"}},"volumeMounts":[{"mountPath":"/home/`+userName+`","name":"home"}]}],"volumes":[{"name":"home","hostPath":{"path":"/home/`+userName+`"}}],"securityContext":{"runAsUser":`+uid+`}}}}}`)
}

func deleteClient(userName, instanceName string) error {
	name := instanceName + "-tensorflow-client"
	return k8s.DeleteReplicationController(userName, name, `{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":0,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"image":"nscc-gz.cn/nscc/tensorflow:1.3.0-cpu-cluster","name":"`+name+`"}]}}}}`)
}

func createParameterServer(userName, instanceName, uid, ip, cpu, memory, parameterServers string) error {
	name := instanceName + "-tensorflow-parameter"
	cwd := "/home/" + userName + "/" + instanceName + "-tensorflow-cpu-cluster"
	return k8s.CreateReplicationController(userName, `{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":`+parameterServers+`,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["tensorflow-parameter.sh","`+ip+`"],"image":"nscc-gz.cn/nscc/tensorflow:1.3.0-cpu-cluster","env":[{"name":"USER","value":"`+userName+`"},{"name":"HOME","value":"/home/`+userName+`"},{"name":"CWD","value":"`+cwd+`"},{"name":"LDAP_SERVER_1","value":"`+k8s.LDAPServer1+`"},{"name":"LDAP_SERVER_2","value":"`+k8s.LDAPServer2+`"},{"name":"TERM","value":"xterm"}],"name":"`+name+`","resources":{"requests":{"cpu":"`+cpu+`m","memory":"`+memory+`Mi"},"limits":{"cpu":"`+cpu+`m","memory":"`+memory+`Mi"}},"volumeMounts":[{"mountPath":"/home/`+userName+`","name":"home"}]}],"volumes":[{"name":"home","hostPath":{"path":"/home/`+userName+`"}}],"securityContext":{"runAsUser":`+uid+`}}}}}`)
}

func deleteParameterServer(userName, instanceName string) error {
	name := instanceName + "-tensorflow-parameter"
	return k8s.DeleteReplicationController(userName, name, `{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":0,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"image":"nscc-gz.cn/nscc/tensorflow:1.3.0-cpu-cluster","name":"`+name+`"}]}}}}`)
}

func createWorkerServer(userName, instanceName, uid, ip, cpu, memory, workerServers string) error {
	name := instanceName + "-tensorflow-worker"
	cwd := "/home/" + userName + "/" + instanceName + "-tensorflow-cpu-cluster"
	return k8s.CreateReplicationController(userName, `{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":`+workerServers+`,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["tensorflow-worker.sh","`+ip+`"],"image":"nscc-gz.cn/nscc/tensorflow:1.3.0-cpu-cluster","env":[{"name":"USER","value":"`+userName+`"},{"name":"HOME","value":"/home/`+userName+`"},{"name":"CWD","value":"`+cwd+`"},{"name":"LDAP_SERVER_1","value":"`+k8s.LDAPServer1+`"},{"name":"LDAP_SERVER_2","value":"`+k8s.LDAPServer2+`"},{"name":"TERM","value":"xterm"}],"name":"`+name+`","resources":{"requests":{"cpu":"`+cpu+`m","memory":"`+memory+`Mi"},"limits":{"cpu":"`+cpu+`m","memory":"`+memory+`Mi"}},"volumeMounts":[{"mountPath":"/home/`+userName+`","name":"home"}]}],"volumes":[{"name":"home","hostPath":{"path":"/home/`+userName+`"}}],"securityContext":{"runAsUser":`+uid+`}}}}}`)
}

func deleteWorkerServer(userName, instanceName string) error {
	name := instanceName + "-tensorflow-worker"
	return k8s.DeleteReplicationController(userName, name, `{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":0,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"image":"nscc-gz.cn/nscc/tensorflow:1.3.0-cpu-cluster","name":"`+name+`"}]}}}}`)
}

func createCluster(userName, instanceName, uid, cpu, memory, parameterServers, workerServers string) (string, error) {
	ip, err := createClientService(userName, instanceName)
	if err != nil {
		return "", err
	}
	err = createClient(userName, instanceName, uid, ip, cpu, memory, parameterServers, workerServers)
	if err != nil {
		deleteClientService(userName, instanceName)
		return "", err
	}
	err = createParameterServer(userName, instanceName, uid, ip, cpu, memory, parameterServers)
	if err != nil {
		deleteClientService(userName, instanceName)
		deleteClient(userName, instanceName)
		return "", err
	}
	err = createWorkerServer(userName, instanceName, uid, ip, cpu, memory, workerServers)
	if err != nil {
		deleteClientService(userName, instanceName)
		deleteClient(userName, instanceName)
		deleteParameterServer(userName, instanceName)
		return "", err
	} else {
		return ip, err
	}
}

func deleteCluster(userName, instanceName string) error {
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
		fmt.Println(`{"message": "usage: ` + os.Args[0] + ` --action create userName instancename uid param"}`)
		os.Exit(1)
	}
	userName := flag.Arg(0)
	instanceName := flag.Arg(1)
	uid := flag.Arg(2)
	data := flag.Arg(3)
	match, _ := regexp.MatchString(`^[a-z0-9][-a-z0-9]*[a-z0-9]$`, userName)
	if !match {
		fmt.Println(`{"message":"userName format error"}`)
		os.Exit(1)
	}
	match, _ = regexp.MatchString(`^[a-z0-9][-a-z0-9]*[a-z0-9]$`, instanceName)
	if !match {
		fmt.Println(`{"message":"instanceName format error"}`)
		os.Exit(1)
	}
	match, _ = regexp.MatchString(`^[1-9][0-9]+`, uid)
	if !match {
		fmt.Println(`{"message":"uid format error"}`)
		os.Exit(1)
	}
	var param Param
	err := json.Unmarshal([]byte(data), &param)
	if err != nil {
		fmt.Println(`{"message":"param format error"}`)
		os.Exit(1)
	}
	if param.CPU == "" || param.Memory == "" || param.ParameterServers == "" || param.WorkerServers == "" {
		fmt.Println(`{"message":"cpu, memory, parameterServers and workerServers are required"}`)
		os.Exit(1)
	}
	ip, err := createCluster(userName, instanceName, uid, param.CPU, param.Memory, param.ParameterServers, param.WorkerServers)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Printf(`{"proxys":[{"proxyName":"tensorflow-ssh","ip":"` + ip + `","port":8000,"protocol":"http","suffix":"","wsSuffix":"ws"},{"proxyName":"jupyter-web","ip":"` + ip + `","port":8080,"protocol":"http","suffix":"api/kernels/","wsSuffix":"api/kernels/"},{"proxyName":"tensorboard","ip":"` + ip + `","port":8888,"protocol":"http","suffix":"","wsSuffix":""}]}`)
	}
}

func delete() {
	if flag.NArg() != 2 {
		fmt.Println(`{"message": "usage: ` + os.Args[0] + ` --action delete userName instancename"}`)
		os.Exit(1)
	}
	userName := flag.Arg(0)
	instanceName := flag.Arg(1)
	match, _ := regexp.MatchString(`^[a-z0-9][-a-z0-9]*[a-z0-9]$`, userName)
	if !match {
		fmt.Println(`{"message":"userName format error"}`)
		os.Exit(1)
	}
	match, _ = regexp.MatchString(`^[a-z0-9][-a-z0-9]*[a-z0-9]$`, instanceName)
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
	fmt.Println(`{"message": "usage: ` + os.Args[0] + ` --action create|delete"}`)
}

func main() {
	action := flag.String("action", "create", "action for manage spark cluster")
	flag.Parse()
	err := k8s.InitK8s()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	switch {
	case *action == "create":
		create()
	case *action == "delete":
		delete()
	default:
		usage()
	}
}
