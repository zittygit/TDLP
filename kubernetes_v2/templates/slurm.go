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
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	Nodes  string `json:"nodes"`
}

func createMasterService(userName, instanceName string) (string, error) {
	name := instanceName + "-slurm-master"
	return k8s.CreateService(userName, `{"apiVersion":"v1","kind":"Service","metadata":{"name":"`+name+`"},"spec":{"ports":[{"name":"server","port":3333,"targetPort":3333},{"name":"ssh","port":8000,"targetPort":8000}],"selector":{"component":"`+name+`"}}}`)
}

func deleteMasterService(userName, instanceName string) error {
	name := instanceName + "-slurm-master"
	return k8s.DeleteService(userName, name)
}

func createMaster(userName, instanceName, uid, ip, cpu, memory, nodes string) error {
	name := instanceName + "-slurm-master"
	cwd := "/home/" + userName + "/" + instanceName + "-slurm"
	return k8s.CreateReplicationController(userName, `{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":1,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["slurmctld.sh","`+ip+`","`+nodes+`"],"image":"nscc-gz.cn/nscc/slurm:17.02.9","env":[{"name":"USER","value":"`+userName+`"},{"name":"HOME","value":"/home/`+userName+`"},{"name":"CWD","value":"`+cwd+`"},{"name":"LDAP_SERVER_1","value":"`+k8s.LDAPServer1+`"},{"name":"LDAP_SERVER_2","value":"`+k8s.LDAPServer2+`"},{"name":"TERM","value":"xterm"}],"name":"`+name+`","resources":{"requests":{"cpu":"`+cpu+`m","memory":"`+memory+`Mi"},"limits":{"cpu":"`+cpu+`m","memory":"`+memory+`Mi"}},"volumeMounts":[{"mountPath":"/home/`+userName+`","name":"home"},{"mountPath":"/dev/gnio0","name":"gn"}],"securityContext":{"privileged":true}}],"volumes":[{"name":"home","hostPath":{"path":"/home/`+userName+`"}},{"hostPath":{"path":"/dev/gnio0"},"name":"gn"}],"securityContext":{"runAsUser":`+uid+`}}}}}`)
}

func deleteMaster(userName, instanceName string) error {
	name := instanceName + "-slurm-master"
	return k8s.DeleteReplicationController(userName, name, `{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":0,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"image":"nscc-gz.cn/nscc/slurm:17.02.9","name":"`+name+`"}]}}}}`)
}

func createCompute(userName, instanceName, uid, ip, cpu, memory, nodes string) error {
	name := instanceName + "-slurm-compute"
	cwd := "/home/" + userName + "/" + instanceName + "-slurm"
	return k8s.CreateReplicationController(userName, `{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":`+nodes+`,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["slurmd.sh","`+ip+`","`+nodes+`"],"image":"nscc-gz.cn/nscc/slurm:17.02.9","env":[{"name":"USER","value":"`+userName+`"},{"name":"HOME","value":"/home/`+userName+`"},{"name":"CWD","value":"`+cwd+`"},{"name":"LDAP_SERVER_1","value":"`+k8s.LDAPServer1+`"},{"name":"LDAP_SERVER_2","value":"`+k8s.LDAPServer2+`"},{"name":"TERM","value":"xterm"}],"name":"`+name+`","resources":{"requests":{"cpu":"`+cpu+`m","memory":"`+memory+`Mi"},"limits":{"cpu":"`+cpu+`m","memory":"`+memory+`Mi"}},"volumeMounts":[{"mountPath":"/home/`+userName+`","name":"home"},{"mountPath":"/dev/gnio0","name":"gn"}],"securityContext":{"runAsUser":`+uid+`,"privileged":true}}],"volumes":[{"name":"home","hostPath":{"path":"/home/`+userName+`"}},{"name":"gn","hostPath":{"path":"/dev/gnio0"}}],"securityContext":{"runAsUser":`+uid+`}}}}}`)
}

func deleteCompute(userName, instanceName string) error {
	name := instanceName + "-slurm-compute"
	return k8s.DeleteReplicationController(userName, name, `{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":0,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"image":"nscc-gz.cn/nscc/slurm:17.02.9","name":"`+name+`"}]}}}}`)
}

func createCluster(userName, instanceName, uid, cpu, memory, nodes string) (string, error) {
	ip, err := createMasterService(userName, instanceName)
	if err != nil {
		return "", err
	}
	err = createMaster(userName, instanceName, uid, ip, cpu, memory, nodes)
	if err != nil {
		deleteMasterService(userName, instanceName)
		return "", err
	}
	err = createCompute(userName, instanceName, uid, ip, cpu, memory, nodes)
	if err != nil {
		deleteMasterService(userName, instanceName)
		deleteMaster(userName, instanceName)
		return "", err
	}
	return ip, err
}

func deleteCluster(userName, instanceName string) error {
	err := deleteCompute(userName, instanceName)
	if err != nil {
		return err
	}
	err = deleteMaster(userName, instanceName)
	if err != nil {
		return err
	}
	return deleteMasterService(userName, instanceName)
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
	if param.CPU == "" || param.Memory == "" || param.Nodes == "" {
		fmt.Println(`{"message":"cpu, memory and nodes are required"}`)
		os.Exit(1)
	}
	ip, err := createCluster(userName, instanceName, uid, param.CPU, param.Memory, param.Nodes)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Printf(`{"proxys":[{"proxyName":"slurm-ssh","ip":"` + ip + `","port":8000,"protocol":"http","suffix":"",wsSuffix":"ws"}]}`)
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
