package app

import (
	"encoding/json"
	"io/ioutil"
	"kubernetes/auth"
	"kubernetes/db"
	"log"
	"net/http"
	"strconv"
)

type App struct {
	Aid     int    `json:"aid"`
	AppName string `json:"appname"`
	Path    string `json:"path"`
	Info    string `json:"info"`
}

func queryApp(w http.ResponseWriter, r *http.Request) {
	var (
		aid     string
		appName string
		path    string
		info    string
	)
	r.ParseForm()
	aid = r.FormValue("aid")
	appName = r.FormValue("appname")
	querystring := "select aid, appname, path, info from app"
	if aid != "" {
		querystring += " where aid=" + aid
	} else {
		if appName != "" {
			querystring += " where appname='" + appName + "'"
		}
	}
	apps, err := db.MysqlDB.Query(querystring)
	defer apps.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query app info"}`))
		return
	}
	str := `{"apps":[`
	for apps.Next() {
		err = apps.Scan(&aid, &appName, &path, &info)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query app info"}`))
			return
		}
		str += `{"aid":` + aid + `,"appname":"` + appName + `","path":"` + path + `","info":"` + info + `"},`
	}
	if str[len(str)-1] == ',' {
		str = str[0:len(str)-1] + "]}"
	} else {
		str += `]}`
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(str))
}

func addApp(w http.ResponseWriter, r *http.Request) {
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var app App
	err := json.Unmarshal(data, &app)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error"}`))
		return
	}
	if app.AppName == "" || app.Path == "" || app.Info == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"appname, path and info are required"}`))
		return
	}
	queryString := "insert app set appname='" + app.AppName + "', path='" + app.Path + "', info='" + app.Info + "'"
	res, err := db.MysqlDB.Exec(queryString)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to add app to mysql"}`))
	} else {
		aid, err := res.LastInsertId()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"failed to add app to mysql"}`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"aid":` + strconv.FormatInt(aid, 10) + `}`))
		}
	}
}

func updateApp(w http.ResponseWriter, r *http.Request) {
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var app App
	err := json.Unmarshal(data, &app)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error"}`))
		return
	}
	if app.Aid == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"aid is required"}`))
		return
	}
	if app.Path != "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"can not change app's path because instance will use it"}`))
		return
	}
	change := false
	queryString := "update app set"
	if app.AppName != "" {
		queryString += " appname='" + app.AppName + "'"
		change = true
	}
	if app.Info != "" {
		if change {
			queryString += ", info='" + app.Info + "'"
		} else {
			queryString += " info='" + app.Info + "'"
			change = true
		}
	}
	queryString += " where aid=" + strconv.Itoa(app.Aid)
	if change {
		_, err := db.MysqlDB.Exec(queryString)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"failed to update app in mysql"}`))
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"app updated"}`))
}

func deleteApp(w http.ResponseWriter, r *http.Request) {
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var app App
	err := json.Unmarshal(data, &app)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error"}`))
		return
	}
	if app.Aid == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"aid is required"}`))
		return
	}
	instances, err := db.MysqlDB.Query("select aid from instance where state=0 and aid=" + strconv.Itoa(app.Aid))
	defer instances.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query instance info"}`))
		return
	}
	if instances.Next() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"running instance of this app exists, can not delete this app"}`))
		return
	}
	_, err = db.MysqlDB.Exec("delete from app where aid=" + strconv.Itoa(app.Aid))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to delete app in mysql"}`))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"app deleted"}`))
	}
}

func AppHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	token, err := auth.JwtAuthRequest(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Access Unauthorized!"}`))
		return
	}
	if r.Method == "GET" {
		queryApp(w, r)
		return
	}
	claim := token.Claims.(*auth.KubernetesClaims)
	if claim.Role != "admin" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"only admin can use this interface!"}`))
		return
	}
	if r.Method == "POST" {
		addApp(w, r)
		return
	}
	if r.Method == "PATCH" {
		updateApp(w, r)
		return
	}
	if r.Method == "DELETE" {
		deleteApp(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"message":"only method GET, POST, PATCH and DELETE are supported"}`))
}
