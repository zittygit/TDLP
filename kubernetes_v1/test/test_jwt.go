package main

import (
	"bytes"
	"fmt"
	"io"
	"kubernetes/auth"
	"log"
	"net/http"
	"time"
)

func main() {
	var (
		uid      int
		userName string
		role     string
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
	req, err := http.NewRequest("GET", "http://10.127.48.18:8080/user", nil)
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
}
