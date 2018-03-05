package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"kubernetes/conf"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func startServer(configFile *string) {
	fmt.Println("starting second proxy server ...")
	myConfig := new(conf.Config)
	myConfig.InitConfig(configFile)
	logFileName := myConfig.Read("logfile")
	if logFileName == "" {
		logFileName = "proxy/second/secondproxy.log"
	}
	pidFileName := myConfig.Read("pidfile")
	if pidFileName == "" {
		pidFileName = "proxy/second/secondproxy.pid"
	}
	pidFile, err := os.OpenFile(pidFileName, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("failed to open pid file: " + pidFileName)
		os.Exit(1)
	}
	defer pidFile.Close()
	d, _, err := bufio.NewReader(pidFile).ReadLine()
	if err != nil {
		if err != io.EOF {
			fmt.Println("failed to read pid file: " + pidFileName)
			os.Exit(1)
		}
	} else {
		pid, err := strconv.Atoi(strings.TrimSpace(string(d)))
		if err != nil {
			fmt.Println("pid format error: " + string(d))
			os.Exit(1)
		}
		if err = syscall.Kill(pid, 0); err == nil {
			fmt.Println("second proxy server with pid " + strconv.Itoa(pid) + " is running, stop it first")
			os.Exit(1)
		}
	}
	cmd := exec.Command("bin/secondproxyd", os.Args[1:]...)
	err = cmd.Start()
	if err != nil {
		fmt.Println("failed to start second proxy server")
		os.Exit(1)
	}
	fmt.Println("second proxy server started with pid " + strconv.Itoa(cmd.Process.Pid))
	fmt.Println("you can view the log in " + logFileName)
	fmt.Println("writing pid to " + pidFileName)
	pidFile.Seek(0, 0)
	pidFile.Truncate(0)
	pidFile.WriteString(strconv.Itoa(cmd.Process.Pid))
}

func stopServer(configFile *string) {
	fmt.Println("stoping second proxy server ...")
	myConfig := new(conf.Config)
	myConfig.InitConfig(configFile)
	logFileName := myConfig.Read("logfile")
	if logFileName == "" {
		logFileName = "proxy/second/secondproxy.log"
	}
	pidFileName := myConfig.Read("pidfile")
	if pidFileName == "" {
		pidFileName = "proxy/second/secondproxy.pid"
	}
	pidFile, err := os.OpenFile(pidFileName, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("failed to open pid file: " + pidFileName)
		os.Exit(1)
	}
	defer pidFile.Close()
	d, _, err := bufio.NewReader(pidFile).ReadLine()
	if err != nil {
		if err != io.EOF {
			fmt.Println("failed to read pid file: " + pidFileName)
			os.Exit(1)
		}
	} else {
		pid, err := strconv.Atoi(strings.TrimSpace(string(d)))
		if err != nil {
			fmt.Println("pid format error: " + string(d))
			os.Exit(1)
		}
		if err = syscall.Kill(pid, 0); err == nil {
			if err = syscall.Kill(pid, syscall.SIGTERM); err == nil {
				for err = syscall.Kill(pid, 0); err == nil; err = syscall.Kill(pid, 0) {
					time.Sleep(time.Second)
				}
				fmt.Println("second proxy server with pid " + strconv.Itoa(pid) + " stopped")
			} else {
				fmt.Println("failed to stop second proxy server with pid " + strconv.Itoa(pid))
				os.Exit(1)
			}
		} else {
			fmt.Println("second proxy server with pid " + strconv.Itoa(pid) + " is not running")
		}
	}
}

func usage() {
	fmt.Println("usage: " + os.Args[0] + " start|stop|restart")
}

func main() {
	configFile := flag.String("config", "proxy/second/secondproxy.ini", "config file for second proxy server")
	flag.Parse()
	action := flag.Arg(0)
	switch {
	case action == "start":
		startServer(configFile)
	case action == "stop":
		stopServer(configFile)
	case action == "restart":
		stopServer(configFile)
		startServer(configFile)
	default:
		usage()
	}
}
