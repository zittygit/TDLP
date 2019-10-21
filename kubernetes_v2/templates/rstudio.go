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
	if param.CPU == "" || param.Memory == "" {
		fmt.Println(`{"message":"cpu and memory are required"}`)
		os.Exit(1)
	}
	name := instanceName + "-rstudio"
	err = k8s.CreateReplicationController(userName, `{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":1,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"command":["start-rserver.sh"],"image":"nscc-gz.cn/nscc/rstudio:1.1.383","env":[{"name":"USER","value":"`+userName+`"},{"name":"HOME","value":"/home/`+userName+`"},{"name":"LDAP_SERVER_1","value":"`+k8s.LDAPServer1+`"},{"name":"LDAP_SERVER_2","value":"`+k8s.LDAPServer2+`"}],"name":"`+name+`","resources":{"requests":{"cpu":"`+param.CPU+`m","memory":"`+param.Memory+`Mi"},"limits":{"cpu":"`+param.CPU+`m","memory":"`+param.Memory+`Mi"}},"volumeMounts":[{"mountPath":"/home/`+userName+`","name":"home"}]}],"volumes":[{"name":"home","hostPath":{"path":"/home/`+userName+`"}}],"securityContext":{"runAsUser":`+uid+`}}}}}`)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ip, err := k8s.CreateService(userName, `{"apiVersion":"v1","kind":"Service","metadata":{"name":"`+name+`"},"spec":{"ports":[{"name":"web","port":8787,"targetPort":8787}],"selector":{"component":"`+name+`"}}}`)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Printf(`{"proxys":[{"proxyName":"rstudio-web","ip":"` + ip + `","port":8787,"protocol":"http","suffix":"",wsSuffix":""}]}`)
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
	name := instanceName + "-rstudio"
	err := k8s.DeleteReplicationController(userName, name, `{"apiVersion":"v1","kind":"ReplicationController","metadata":{"name":"`+name+`"},"spec":{"replicas":0,"selector":{"component":"`+name+`"},"template":{"metadata":{"labels":{"component":"`+name+`"}},"spec":{"containers":[{"image":"nscc-gz.cn/nscc/rstudio:1.1.383","name":"`+name+`"}]}}}}`)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = k8s.DeleteService(userName, name)
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