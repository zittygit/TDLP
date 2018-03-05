package main

import (
	"flag"
	"fmt"
	"kubernetes/conf"
	"kubernetes/k8s"
)

func main() {
	configFile := flag.String("config", "k8s/k8s.ini", "config file for kubernetes")
	flag.Parse()
	myConfig := new(conf.Config)
	myConfig.InitConfig(configFile)
	k8s.BearerToken = myConfig.Read("bearertoken")
	if k8s.BearerToken == "" {
		k8s.BearerToken = "Bearer 6B1GbqhcjqGYPAAy285otYhUUV4z4kiu"
	}
	k8s.K8sApiServer = myConfig.Read("k8sapiserver")
	if k8s.K8sApiServer == "" {
		k8s.K8sApiServer = "https://10.127.48.18:6443/api/v1"
	}
	err := k8s.CreateNameSpace("test")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("namespace test added")
	}
	err = k8s.DeleteNameSpace("test")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("namespace test deleted")
	}
}
