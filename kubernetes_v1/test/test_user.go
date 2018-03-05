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
		userName string
		role     string
		passWord string
		gid      string
		uid      string
		email    string
	)
	fmt.Print("Please input uid: ")
	fmt.Scan(&uid)
	fmt.Print("Please input username: ")
	fmt.Scan(&userName)
	fmt.Print("Please input role: ")
	fmt.Scan(&role)
	auth.JwtInit("auth/kubernetes.rsa", "auth/kubernetes.rsa.pub")
	id, _ := strconv.Atoi(uid)
	token, err := auth.JwtCreateToken(id, userName, role)
	if err != nil {
		fmt.Println("error while Signing Token!")
		return
	}
	fmt.Println("Authorize with token: " + token)
	fmt.Print("Please input username: ")
	fmt.Scan(&userName)
	fmt.Print("Please input password: ")
	fmt.Scan(&passWord)
	req, err := http.NewRequest("POST", "http://10.127.48.18:8080/user", strings.NewReader(`{"username":"`+userName+`","password":"`+passWord+`"}`))
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
	req, err = http.NewRequest("GET", "http://10.127.48.18:8080/user", nil)
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
	fmt.Print("Please input uid: ")
	fmt.Scan(&uid)
	fmt.Print("Please input gid: ")
	fmt.Scan(&gid)
	fmt.Print("Please input role: ")
	fmt.Scan(&role)
	fmt.Print("Please input email: ")
	fmt.Scan(&email)
	req, err = http.NewRequest("PATCH", "http://10.127.48.18:8080/user", strings.NewReader(`{"uid":`+uid+`,"gid":`+gid+`,"role":"`+role+`","email":"`+email+`"}`))
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
	fmt.Print("Please input uid: ")
	fmt.Scan(&uid)
	req, err = http.NewRequest("GET", "http://10.127.48.18:8080/user?uid="+uid, nil)
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
	fmt.Print("Please input username: ")
	fmt.Scan(&userName)
	req, err = http.NewRequest("GET", "http://10.127.48.18:8080/user?username="+userName, nil)
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
	req, err = http.NewRequest("GET", "http://10.127.48.18:8080/user?gid="+gid, nil)
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
	fmt.Print("Please input uid: ")
	fmt.Scan(&uid)
	req, err = http.NewRequest("DELETE", "http://10.127.48.18:8080/user", strings.NewReader(`{"uid":`+uid+`}`))
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
