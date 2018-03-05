package main

import (
	"bytes"
	"fmt"
	"io"
	"kubernetes/auth"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	var (
		uid          int
		userName     string
		role         string
		iid          string
		instanceName string
		cpu          string
		memory       string
		nodes        string
		param        string
	)
	fmt.Println("Generate token")
	fmt.Print("Please input uid: ")
	fmt.Scan(&uid)
	fmt.Print("Please input username: ")
	fmt.Scan(&userName)
	fmt.Print("Please input role: ")
	fmt.Scan(&role)
	auth.JwtInit("auth/kubernetes.rsa", "auth/kubernetes.rsa.pub")
	token, err := auth.JwtCreateToken(uid, userName, role)
	if err != nil {
		fmt.Println("error while Signing Token!")
		return
	}
	fmt.Println("Authorize with token: " + token)
	fmt.Println("Create instance")
	fmt.Print("Please input instanceName: ")
	fmt.Scan(&instanceName)
	fmt.Print("Please input cpu: ")
	fmt.Scan(&cpu)
	fmt.Print("Please input memory: ")
	fmt.Scan(&memory)
	fmt.Print("Please input nodes: ")
	fmt.Scan(&nodes)
	param = `{"cpu":` + cpu + `,"memory":` + memory + `,"nodes":` + nodes + `}`
	req, err := http.NewRequest("POST", "http://10.127.48.18:8080/instance", strings.NewReader(`{"aid":1,"instancename":"`+instanceName+`","param":`+strconv.Quote(param)+`}`))
	if err != nil {
		log.Fatal(err)
		return
	}
	req.AddCookie(&http.Cookie{Name: "kubernetes_token", Value: token, Expires: time.Now().Add(time.Hour * 24)})
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	buf := new(bytes.Buffer)
	io.Copy(buf, res.Body)
	res.Body.Close()
	fmt.Println(buf.String())
	fmt.Println("Query all instances")
	req, err = http.NewRequest("GET", "http://10.127.48.18:8080/instance?kind=all", nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	req.AddCookie(&http.Cookie{Name: "kubernetes_token", Value: token, Expires: time.Now().Add(time.Hour * 24)})
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	buf = new(bytes.Buffer)
	io.Copy(buf, res.Body)
	res.Body.Close()
	fmt.Println(buf.String())
	fmt.Println("Delete instances use iid")
	fmt.Printf("Please input iid: ")
	fmt.Scanf("%s", &iid)
	req, err = http.NewRequest("DELETE", "http://10.127.48.18:8080/instance", strings.NewReader(`{"iid":`+iid+`}`))
	if err != nil {
		log.Fatal(err)
		return
	}
	req.AddCookie(&http.Cookie{Name: "kubernetes_token", Value: token, Expires: time.Now().Add(time.Hour * 24)})
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	buf = new(bytes.Buffer)
	io.Copy(buf, res.Body)
	res.Body.Close()
	fmt.Println(buf.String())
	fmt.Println("Query all instances use instancename")
	fmt.Print("Please input instancename: ")
	fmt.Scan(&instanceName)
	req, err = http.NewRequest("GET", "http://10.127.48.18:8080/instance?kind=all&&instancename="+instanceName, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	req.AddCookie(&http.Cookie{Name: "kubernetes_token", Value: token, Expires: time.Now().Add(time.Hour * 24)})
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	buf = new(bytes.Buffer)
	io.Copy(buf, res.Body)
	res.Body.Close()
	fmt.Println(buf.String())
	fmt.Println("Query all running instances use instancename")
	fmt.Print("Please input instancename: ")
	fmt.Scan(&instanceName)
	req, err = http.NewRequest("GET", "http://10.127.48.18:8080/instance?kind=all&&instancename="+instanceName+"&&state=0", nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	req.AddCookie(&http.Cookie{Name: "kubernetes_token", Value: token, Expires: time.Now().Add(time.Hour * 24)})
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	buf = new(bytes.Buffer)
	io.Copy(buf, res.Body)
	res.Body.Close()
	fmt.Println(buf.String())
	fmt.Println("Query all deleted instances use instancename")
	fmt.Print("Please input instancename: ")
	fmt.Scan(&instanceName)
	req, err = http.NewRequest("GET", "http://10.127.48.18:8080/instance?kind=all&&instancename="+instanceName+"&&state=1", nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	req.AddCookie(&http.Cookie{Name: "kubernetes_token", Value: token, Expires: time.Now().Add(time.Hour * 24)})
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	buf = new(bytes.Buffer)
	io.Copy(buf, res.Body)
	res.Body.Close()
	fmt.Println(buf.String())
	fmt.Println("Update instance")
	fmt.Print("Please input iid: ")
	fmt.Scan(&iid)
	fmt.Print("Please input cpu: ")
	fmt.Scan(&cpu)
	fmt.Print("Please input memory: ")
	fmt.Scan(&memory)
	fmt.Print("Please input nodes: ")
	fmt.Scan(&nodes)
	param = `{"cpu":` + cpu + `,"memory":` + memory + `,"nodes":` + nodes + `}`
	req, err = http.NewRequest("PATCH", "http://10.127.48.18:8080/instance", strings.NewReader(`{"aid":1,"iid":`+iid+`,"param":`+strconv.Quote(param)+`}`))
	if err != nil {
		log.Fatal(err)
		return
	}
	req.AddCookie(&http.Cookie{Name: "kubernetes_token", Value: token, Expires: time.Now().Add(time.Hour * 24)})
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	buf = new(bytes.Buffer)
	io.Copy(buf, res.Body)
	res.Body.Close()
	fmt.Println(buf.String())
	fmt.Println("Query single instance use iid")
	fmt.Print("Please input iid: ")
	fmt.Scan(&iid)
	req, err = http.NewRequest("GET", "http://10.127.48.18:8080/instance?kind=single&&iid="+iid, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	req.AddCookie(&http.Cookie{Name: "kubernetes_token", Value: token, Expires: time.Now().Add(time.Hour * 24)})
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	buf = new(bytes.Buffer)
	io.Copy(buf, res.Body)
	fmt.Println(buf.String())
	fmt.Println("Query proxy information of running instance use iid")
	fmt.Print("Please input iid: ")
	fmt.Scan(&iid)
	req, err = http.NewRequest("GET", "http://10.127.48.18:8080/instance?kind=proxy&&iid="+iid, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	req.AddCookie(&http.Cookie{Name: "kubernetes_token", Value: token, Expires: time.Now().Add(time.Hour * 24)})
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	buf = new(bytes.Buffer)
	io.Copy(buf, res.Body)
	fmt.Println(buf.String())
}
