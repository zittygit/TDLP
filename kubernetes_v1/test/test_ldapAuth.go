package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func main() {
	var (
		userName string
		passWord string
	)
	fmt.Printf("Please input your username: ")
	fmt.Scanf("%s", &userName)
	fmt.Printf("Please input your password: ")
	fmt.Scanf("%s", &passWord)
	req, err := http.NewRequest("POST", "http://10.127.48.18:8080/auth", strings.NewReader(`{"username":"`+userName+`","password":"`+passWord+`"}`))
	if err != nil {
		log.Fatal(err)
		return
	}
	res, err := http.DefaultClient.Do(req)
	fmt.Println(res.Header)
	if err != nil {
		log.Fatal(err)
		return
	}
	buf := new(bytes.Buffer)
	io.Copy(buf, res.Body)
	res.Body.Close()
	fmt.Println(buf.String())
}
