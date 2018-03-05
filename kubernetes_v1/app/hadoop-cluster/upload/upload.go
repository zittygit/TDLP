package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"encoding/json"
)
func checkErr(err error) {
	if err != nil {
		err.Error()
	}
}
func upload(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "GET" {
		t, err := template.ParseFiles("upload.gptl")
		checkErr(err)
		t.Execute(w, nil)
	} else {
		file, handle, err := r.FormFile("file")
		checkErr(err)
		f, err := os.OpenFile("/root/input/"+handle.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		io.Copy(f, file)
		checkErr(err)
		defer f.Close()
		defer file.Close()
		fmt.Println("upload success")
	}
}
func getFileList(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(getFileJson("/root/output/")))
}
func getFileJson(path string) string {
	type File struct {
		FileName     string  `json:"name"`
		Size         string  `json:"size"`
	}
	var files  []File
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		file := File{FileName:f.Name(),Size:fmt.Sprint(f.Size()/1024)+"KB"}
		files = append(files, file)
		return nil
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
		return ""
	}
	data, err := json.Marshal(files)
	if err == nil {
		return string(data)
	}
	return ""
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("/bin/upload/index.html")
	if err != nil {
		w.Write([]byte("parse template error: " + err.Error()))
		return
	}
	t.Execute(w, nil)
}

func mkdir(path string) error {
	err := os.MkdirAll(path, 0666)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	//创建上传下载目录
	err := mkdir("/root/input")
	if err != nil{
		fmt.Println(err)
		return
	}
	err = mkdir("/root/output")
	if err != nil{
		fmt.Println(err)
		return
	}
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/fileList",getFileList)
	http.HandleFunc("/download/",func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/root/output/"+r.URL.Path[10:])
	})
	http.HandleFunc("/js/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/bin/upload/"+r.URL.Path[4:])
	})
	http.HandleFunc("/", loginHandler)
	err = http.ListenAndServe(":8001", nil)
	if err != nil {
		log.Fatal("listenAndServe: ", err)
	}

}
