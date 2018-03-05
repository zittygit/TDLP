package main

import (
	"bytes"
	"fmt"
	"io"
	"kubernetes/auth"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	var (
		uid       int
		userName  string
		role      string
		groupName string
		gid       string
	)
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
	fmt.Print("Please input groupname: ")
	fmt.Scan(&groupName)
	req, err := http.NewRequest("POST", "http://10.127.48.18:8080/group", strings.NewReader(`{"groupname":"`+groupName+`"}`))
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
	req, err = http.NewRequest("GET", "http://10.127.48.18:8080/group", nil)
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
	fmt.Print("Please input groupname: ")
	fmt.Scan(&groupName)
	req, err = http.NewRequest("GET", "http://10.127.48.18:8080/group?groupname="+groupName, nil)
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
	fmt.Print("Please input gid: ")
	fmt.Scan(&gid)
	req, err = http.NewRequest("GET", "http://10.127.48.18:8080/group?gid="+gid, nil)
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
	fmt.Print("Please input gid: ")
	fmt.Scan(&gid)
	req, err = http.NewRequest("DELETE", "http://10.127.48.18:8080/group", strings.NewReader(`{"gid":`+gid+`}`))
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
}
