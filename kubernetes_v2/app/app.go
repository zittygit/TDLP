package app

import (
	"encoding/json"
	"io/ioutil"
	"kubernetes/auth"
	"kubernetes/db"
	"kubernetes/user"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

var ProxyAddr string

const (
	RUNNING  = 0
	FINISHED = 1
)

type App struct {
	Aid     int    `json:"aid"`
	AppName string `json:"appName"`
	Tid     int    `json:"tid"`
	Param   string `json:"param"`
}

type Proxy struct {
	ProxyName string `json:"proxyName"`
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	Protocol  string `json:"protocol"`
	Suffix    string `json:"suffix"`
	WSSuffix  string `json:"wsSuffix"`
}

type ProxyList struct {
	Proxys []Proxy `json:"proxys"`
}

func queryApp(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	var (
		kind         string
		aid          string
		appName      string
		templateName string
		state        string
		param        string
		createTime   string
		deleteTime   *string
		pid          string
		proxy        Proxy
	)
	r.ParseForm()
	kind = r.FormValue("kind")
	switch {
	case kind == "app":
		{
			query := "select aid, appName, templateName, state, apps.param, createTime, deleteTime from apps, templates"
			cond := false
			if claim.Role != user.ADMIN {
				query += " where uid=" + strconv.Itoa(claim.Uid)
				cond = true
			}
			aid = r.FormValue("aid")
			if aid != "" {
				_, err := strconv.Atoi(aid)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"message":"aid must be integer!"}`))
					return
				}
				if cond {
					query += " and aid=" + aid
				} else {
					query += " where aid=" + aid
					cond = true
				}
			}
			appName = r.FormValue("appName")
			if appName != "" {
				if cond {
					query += " and appName='" + appName + "'"
				} else {
					query += " where appName='" + appName + "'"
					cond = true
				}
			}
			state = r.FormValue("state")
			if state != "" {
				if cond {
					query += " and state=" + state
				} else {
					query += " where state=" + state
				}
			}
			if cond {
				query += " and apps.tid=templates.tid"
			} else {
				query += " where apps.tid=templates.tid"
			}
			apps, err := db.MysqlDB.Query(query)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query app info!"}`))
				return
			}
			defer apps.Close()
			str := `{"apps":[`
			for apps.Next() {
				err = apps.Scan(&aid, &appName, &templateName, &state, &param, &createTime, &deleteTime)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"failed to query app info!"}`))
					return
				}
				if deleteTime == nil {
					deleteTime = new(string)
					*deleteTime = ""
				}
				str += `{"aid":` + aid + `,"appName":"` + appName + `","templateName":"` + templateName + `","state":` + state + `,"param":` + strconv.Quote(param) + `,"createTime":"` + createTime + `","deleteTime":"` + *deleteTime + `"},`
			}
			if str[len(str)-1] == ',' {
				str = str[0:len(str)-1] + "]}"
			} else {
				str += `]}`
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(str))
		}
	case kind == "proxy":
		{
			aid = r.FormValue("aid")
			if aid == "" {
				log.Println("aid is required!")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"message":"aid is required!"}`))
				return
			}
			_, err := strconv.Atoi(aid)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"message":"aid must be integer!"}`))
				return
			}
			proxys, err := db.MysqlDB.Query("select pid from proxys where aid=" + aid)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query proxy info!"}`))
				return
			}
			defer proxys.Close()
			token, err := r.Cookie("kubernetes_token")
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"internal server error!"}`))
				return
			}
			str := `{"proxys":[`
			for proxys.Next() {
				err = proxys.Scan(&pid)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"failed to query proxy info!"}`))
					return
				}
				req, err := http.NewRequest("GET", ProxyAddr+"/?pid="+pid, nil)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"internal server error!"}`))
					return
				}
				req.AddCookie(token)
				res, err := http.DefaultClient.Do(req)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"internal server error!"}`))
					return
				}
				data, _ := ioutil.ReadAll(res.Body)
				res.Body.Close()
				if res.StatusCode != http.StatusOK {
					log.Println(string(data))
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"internal server error!"}`))
					return
				}
				err = json.Unmarshal(data, &proxy)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"internal server error!"}`))
					return
				}
				if proxy.Protocol == "http" {
					str += `{"proxyName":"` + proxy.ProxyName + `","url":"http://` + proxy.IP + `:` + strconv.Itoa(proxy.Port) + `"}`
				} else {
					str += `{"proxyName":"` + proxy.ProxyName + `","url":"https://` + proxy.IP + `:` + strconv.Itoa(proxy.Port) + `"}`
				}
			}
			if str[len(str)-1] == ',' {
				str = str[0:len(str)-1] + "]}"
			} else {
				str += `]}`
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(str))
		}
	default:
		{
			log.Println("kind must be app or proxy!")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"kind must be app or proxy!"}`))
		}
	}
}

func createApp(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	var (
		app       App
		path      string
		proxyList ProxyList
	)
	active := user.CheckActive(w, r, claim)
	if active == user.INACTIVE {
		return
	}
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	err := json.Unmarshal(data, &app)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error!"}`))
		return
	}
	if app.AppName == "" || app.Tid == 0 || app.Param == "" {
		log.Println("appName, tid and param are required!")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"appName, tid and param are required!"}`))
		return
	}
	req, err := http.NewRequest("GET", ProxyAddr+"/health", nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"message":"Internal Server Error!"`))
		return
	}
	res, err := http.DefaultClient.Do(req)
	res.Body.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"message":"proxy server not running!"`))
		return
	}
	apps, err := db.MysqlDB.Query("select appName from apps where uid=" + strconv.Itoa(claim.Uid) + " and tid=" + strconv.Itoa(app.Tid) + " and appName='" + app.AppName + "' and state=0")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query app info!"}`))
		return
	}
	defer apps.Close()
	if apps.Next() {
		log.Println("running app with name " + app.AppName + " already exists!")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"running app with name ` + app.AppName + ` already exists!"}`))
		return
	}
	templates, err := db.MysqlDB.Query("select path from templates where tid=" + strconv.Itoa(app.Tid))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query template info!"}`))
		return
	}
	defer templates.Close()
	if templates.Next() {
		err = templates.Scan(&path)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query template info!"}`))
			return
		}
	} else {
		log.Println("template not found!")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"template not found!"}`))
		return
	}
	cmd := exec.Command(path, "--action", "create", claim.UserName, app.AppName, strconv.Itoa(claim.Uid), app.Param)
	buf, err := cmd.Output()
	if err != nil {
		log.Println(err)
		log.Println(string(buf))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"app create failed!"}`))
		return
	}
	err = json.Unmarshal(buf, &proxyList)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"app create failed!"}`))
		return
	}
	result, err := db.MysqlDB.Exec("insert apps set appName='" + app.AppName + "', tid=" + strconv.Itoa(app.Tid) + ", uid=" + strconv.Itoa(claim.Uid) + ", param=" + strconv.Quote(app.Param) + ", createTime=now()")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"app creation failed!"}`))
		return
	}
	aid, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"instance creation failed"}`))
		return
	}
	token, err := r.Cookie("kubernetes_token")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal server error!"}`))
		return
	}
	for _, proxy := range proxyList.Proxys {
		result, err = db.MysqlDB.Exec("insert proxys set proxyName='" + proxy.ProxyName + "', aid=" + strconv.FormatInt(aid, 10) + ", ip='" + proxy.IP + "', port=" + strconv.Itoa(proxy.Port) + ", protocol='" + proxy.Protocol + "', suffix='" + proxy.Suffix + "', wsSuffix='" + proxy.WSSuffix + "'")
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"app's proxy creation failed!"}`))
			return
		}
		pid, err := result.LastInsertId()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"app's proxy creation failed!"}`))
			return
		}
		req, err = http.NewRequest("POST", ProxyAddr, strings.NewReader(`{"pid":`+strconv.FormatInt(pid, 10)+`}`))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"Internal Server Error!"`))
			return
		}
		req.AddCookie(token)
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"proxy server not correct running!"`))
			return
		}
		if res.StatusCode != http.StatusCreated {
			data, _ = ioutil.ReadAll(res.Body)
			res.Body.Close()
			log.Println(string(data))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"proxy server not correct running!"`))
			return
		}
		res.Body.Close()
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"aid":` + strconv.FormatInt(aid, 10) + `}`))
}

func deleteApp(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	var (
		app     App
		path    string
		appName string
		pid     string
	)
	active := user.CheckActive(w, r, claim)
	if active == user.INACTIVE {
		return
	}
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	err := json.Unmarshal(data, &app)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error!"}`))
		return
	}
	if app.Aid == 0 {
		log.Println("aid is required!")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"aid is required!"}`))
		return
	}
	req, err := http.NewRequest("GET", ProxyAddr+"/health", nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"message":"Internal Server Error!"`))
		return
	}
	res, err := http.DefaultClient.Do(req)
	res.Body.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"message":"proxy server not running!"`))
		return
	}
	apps, err := db.MysqlDB.Query("select path, appName from templates, apps where aid=" + strconv.Itoa(app.Aid) + " and state=0 and apps.tid=templates.tid")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query app info!"}`))
		return
	}
	defer apps.Close()
	if apps.Next() {
		err = apps.Scan(&path, &appName)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query app info!"}`))
			return
		}
	} else {
		log.Println("app not found!")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"app not found!"}`))
		return
	}
	cmd := exec.Command(path, "--action", "delete", claim.UserName, appName)
	buf, err := cmd.Output()
	if err != nil {
		log.Println(err)
		log.Println(string(buf))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"app delete failed!"}`))
		return
	}
	_, err = db.MysqlDB.Exec("update apps set state=1, deleteTime=now() where aid=" + strconv.Itoa(app.Aid))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to update app in mysql!"}`))
		return
	}
	token, err := r.Cookie("kubernetes_token")
	proxys, err := db.MysqlDB.Query("select pid from proxys where aid=" + strconv.Itoa(app.Aid))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query proxy info!"}`))
		return
	}
	defer proxys.Close()
	for proxys.Next() {
		err = proxys.Scan(&pid)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query proxy info!"}`))
			return
		}
		req, err := http.NewRequest("DELETE", ProxyAddr, strings.NewReader(`{"pid":`+pid+`}`))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"Internal Server Error!"`))
			return
		}
		req.AddCookie(token)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"proxy server not correct running!"`))
			return
		}
		if res.StatusCode != http.StatusOK {
			data, _ = ioutil.ReadAll(res.Body)
			res.Body.Close()
			log.Println(string(data))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"proxy server not correct running!"`))
			return
		}
		res.Body.Close()
	}
	_, err = db.MysqlDB.Exec("delete from proxys where aid=" + strconv.Itoa(app.Aid))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to delete proxy info!"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"instance delete successful!"}`))
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
	claim := token.Claims.(*auth.KubernetesClaims)
	if r.Method == "GET" {
		queryApp(w, r, claim)
		return
	}
	if r.Method == "POST" {
		createApp(w, r, claim)
		return
	}
	if r.Method == "DELETE" {
		deleteApp(w, r, claim)
		return
	}
	log.Println("only method GET, POST and DELETE are supported")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"message":"only method GET, POST and DELETE are supported"}`))
}
