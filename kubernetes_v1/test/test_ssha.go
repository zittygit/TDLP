package main

import (
	"fmt"
	"kubernetes/auth"
	"log"
	"os/exec"
)

func main() {
	s, _ := auth.GenerateSSHA("guoguixin")
	fmt.Println(s)
	valid := auth.ValidateSSHA("guoguixin", s)
	if valid {
		fmt.Println("validate successful")
	} else {
		fmt.Println("not validate")
	}
	cmd := exec.Command("slappasswd", "-n", "-s", "guoguixin")
	buf, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	s = string(buf)
	fmt.Println(s)
	valid = auth.ValidateSSHA("guoguixin", s)
	if valid {
		fmt.Println("validate successful")
	} else {
		fmt.Println("not validate")
	}
}
