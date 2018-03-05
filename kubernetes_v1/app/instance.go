package app

import (
	"encoding/json"
	"io/ioutil"
	"kubernetes/auth"
	"kubernetes/db"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Instance struct {
	Iid          int    `json:"iid"`
	InstanceName string `json:"instancename"`
	Aid          int    `json:"aid"`
	Param        string `json:"param"`
}

type Proxy struct {
	ProxyName    string `json:"proxyname"`
	HttpUrl      string `json:"httpurl"`
	WebSocketUrl string `json:"websocketurl"`
}

type ProxyList struct {
	Services []Proxy `json:"services"`
}

var (
	FirstProxyAddr  string
	SecondProxyAddr string
	SecondProxyIp   string
)

func queryInstance(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	var (
		kind         string
		iid          string
		instanceName string
		aid          *string
		appName      string
		uid          string
		userName     string
		param        string
		createTime   string
		deleteTime   *string
		state        string
		proxyName    string
		secondPort   string
	)
	r.ParseForm()
	kind = r.FormValue("kind")
	switch {
	case kind == "all":
		{
			queryString := "select iid, instancename, aid, uid, state, createtime, deletetime from instance"
			cond := false
			if claim.Role != "admin" {
				queryString += " where uid=" + strconv.Itoa(claim.Uid)
				cond = true
			}
			instanceName = r.FormValue("instancename")
			if instanceName != "" {
				if cond {
					queryString += " and instancename='" + instanceName + "'"
				} else {
					queryString += " where instancename='" + instanceName + "'"
					cond = true
				}
			}
			state = r.FormValue("state")
			if state != "" {
				if cond {
					queryString += " and state=" + state
				} else {
					queryString += " where state=" + state
				}
			}
			instances, err := db.MysqlDB.Query(queryString)
			defer instances.Close()
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query instance info"}`))
				return
			}
			str := `{"instances":[`
			for instances.Next() {
				err = instances.Scan(&iid, &instanceName, &aid, &uid, &state, &createTime, &deleteTime)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"failed to query instance info"}`))
					return
				}
				if aid == nil {
					aid = new(string)
					*aid = "0"
					appName = ""
				} else {
					app, err := db.MysqlDB.Query("select appname from app where aid=" + *aid)
					defer app.Close()
					if err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(`{"message":"failed to query app info"}`))
						return
					}
					if app.Next() {
						err = app.Scan(&appName)
						if err != nil {
							log.Println(err)
							w.WriteHeader(http.StatusInternalServerError)
							w.Write([]byte(`{"message":"failed to query app info"}`))
							return
						}
					} else {
						appName = ""
					}
				}
				user, err := db.MysqlDB.Query("select username from user where uid=" + uid)
				defer user.Close()
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"failed to query user info"}`))
					return
				}
				if user.Next() {
					err = user.Scan(&userName)
					if err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(`{"message":"failed to query user info"}`))
						return
					}
				} else {
					userName = ""
				}
				if deleteTime == nil {
					deleteTime = new(string)
					*deleteTime = ""
				}
				str += `{"iid":` + iid + `,"instancename":"` + instanceName + `","aid":` + *aid + `,"appname":"` + appName + `","uid":` + uid + `,"username":"` + userName + `","state":` + state + `,"createtime":"` + createTime + `","deletetime":"` + *deleteTime + `"},`
			}
			if str[len(str)-1] == ',' {
				str = str[0:len(str)-1] + "]}"
			} else {
				str += `]}`
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(str))
		}
	case kind == "single":
		{
			iid = r.FormValue("iid")
			if iid == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"message":"iid is required"}`))
				return
			}
			queryString := "select iid, instancename, aid, createtime, deletetime from instance where iid=" + iid
			if claim.Role != "admin" {
				queryString += " and uid=" + strconv.Itoa(claim.Uid)
			}
			instance, err := db.MysqlDB.Query(queryString)
			defer instance.Close()
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query instance info"}`))
				return
			}
			if instance.Next() {
				str := `{"instance":`
				err = instance.Scan(&iid, &instanceName, &aid, &createTime, &deleteTime)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"failed to query instance info"}`))
					return
				}
				if aid == nil {
					aid = new(string)
					*aid = "0"
					appName = ""
				} else {
					app, err := db.MysqlDB.Query("select appname from app where aid=" + *aid)
					defer app.Close()
					if err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(`{"message":"failed to query app info"}`))
						return
					}
					if app.Next() {
						err = app.Scan(&appName)
						if err != nil {
							log.Println(err)
							w.WriteHeader(http.StatusInternalServerError)
							w.Write([]byte(`{"message":"failed to query app info"}`))
							return
						}
					} else {
						appName = ""
					}
				}
				if deleteTime == nil {
					deleteTime = new(string)
					*deleteTime = ""
				}
				str += `{"iid":` + iid + `,"instancename":"` + instanceName + `","aid":` + *aid + `,"appname":"` + appName + `","uid":` + strconv.Itoa(claim.Uid) + `,"username":"` + claim.UserName + `","createtime":"` + createTime + `","deletetime":"` + *deleteTime + `"},"config":[`
				config, err := db.MysqlDB.Query("select starttime, endtime, param from config where iid=" + iid + " order by starttime desc")
				defer config.Close()
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"failed to query instance's config info"}`))
					return
				}
				for config.Next() {
					err = config.Scan(&createTime, &deleteTime, &param)
					if err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(`{"message":"failed to query instance's config info"}`))
						return
					}
					if deleteTime == nil {
						deleteTime = new(string)
						*deleteTime = ""
					}
					str += `{"starttime":"` + createTime + `","endtime":"` + *deleteTime + `","param":` + strconv.Quote(param) + `},`
				}
				if str[len(str)-1] == ',' {
					str = str[0:len(str)-1] + "]}"
				} else {
					str += `]}`
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(str))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"instance":{}}`))
		}
	case kind == "proxy":
		{
			iid = r.FormValue("iid")
			if iid == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"message":"iid is required"}`))
				return
			}
			proxys, err := db.MysqlDB.Query("select proxyname, secondport from proxy where iid=" + iid)
			defer proxys.Close()
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query proxy info"}`))
				return
			}
			str := `{"Services":[`
			for proxys.Next() {
				err = proxys.Scan(&proxyName, &secondPort)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"failed to query proxy info"}`))
					return
				}
				str += `{"proxyname":"` + proxyName + `","proxyurl":"http://` + SecondProxyIp + `:` + secondPort + `"},`
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
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"kind must be all, single or proxy"}`))
		}
	}
}

func addInstance(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var instance Instance
	err := json.Unmarshal(data, &instance)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error"}`))
		return
	}
	if instance.InstanceName == "" || instance.Aid == 0 || instance.Param == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"instancename, aid and param are required"}`))
		return
	}
	req, err := http.NewRequest("CHECK", FirstProxyAddr, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"message":"Internal Server Error!"`))
		return
	}
	tokenCookie, err := r.Cookie("kubernetes_token")
	req.AddCookie(&http.Cookie{Name: "kubernetes_token", Value: tokenCookie.Value, Expires: time.Now().Add(time.Hour * 24)})
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"message":"first proxy server not running"`))
		return
	}
	req, err = http.NewRequest("CHECK", SecondProxyAddr, nil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"message":"Internal Server Error!"`))
		return
	}
	req.AddCookie(&http.Cookie{Name: "kubernetes_token", Value: tokenCookie.Value, Expires: time.Now().Add(time.Hour * 24)})
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"message":"second proxy server not running"`))
		return
	}
	instances, err := db.MysqlDB.Query("select instancename from instance where instancename='" + instance.InstanceName + "' and state=0")
	defer instances.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query instance info"}`))
		return
	}
	if instances.Next() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"running instance with this name already exists"}`))
		return
	}
	app, err := db.MysqlDB.Query("select path from app where aid=" + strconv.Itoa(instance.Aid))
	defer app.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query app info"}`))
		return
	}
	var path string
	if app.Next() {
		err = app.Scan(&path)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query app info"}`))
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"aid not exists in mysql"}`))
		return
	}
	cmd := exec.Command(path, "--action", "create", strconv.Itoa(claim.Uid), claim.UserName, instance.InstanceName, instance.Param)
	buf, err := cmd.Output()
	if err != nil {
		log.Println(err)
		log.Println(string(buf))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"instance creation failed"}`))
		return
	}
	var proxyList ProxyList
	err = json.Unmarshal(buf, &proxyList)
	if err != nil {
		log.Println("json format error")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"instance creation failed"}`))
		return
	}
	res, err := db.MysqlDB.Exec("insert instance set instancename='" + instance.InstanceName + "', aid=" + strconv.Itoa(instance.Aid) + ", uid=" + strconv.Itoa(claim.Uid) + ", createtime=now()")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"instance creation failed"}`))
		return
	}
	iid, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"instance creation failed"}`))
		return
	}
	res, err = db.MysqlDB.Exec("insert config set iid=" + strconv.FormatInt(iid, 10) + ", starttime=now(), param='" + instance.Param + "'")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"instance's config creation failed"}`))
		return
	}
	cid, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"instance's config creation failed"}`))
		return
	}
	_, err = db.MysqlDB.Exec("update instance set cid=" + strconv.FormatInt(cid, 10) + " where iid=" + strconv.FormatInt(iid, 10))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"instance's update failed"}`))
		return
	}
	for _, proxy := range proxyList.Services {
		res, err = db.MysqlDB.Exec("insert proxy set proxyname='" + proxy.ProxyName + "', iid=" + strconv.FormatInt(iid, 10) + ", firstport=0, secondport=0, httpurl='" + proxy.HttpUrl + "', websocketurl='" + proxy.WebSocketUrl + "'")
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"instance's proxy creation failed"}`))
			return
		}
		pid, err := res.LastInsertId()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"instance's proxy creation failed"}`))
			return
		}
		req, err = http.NewRequest("POST", FirstProxyAddr, strings.NewReader(`{"pid":`+strconv.FormatInt(pid, 10)+`}`))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"Internal Server Error"`))
			return
		}
		req.AddCookie(&http.Cookie{Name: "kubernetes_token", Value: tokenCookie.Value, Expires: time.Now().Add(time.Hour * 24)})
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"first proxy server not correct running"`))
			return
		}
		if res.StatusCode != http.StatusCreated {
			data, _ = ioutil.ReadAll(res.Body)
			res.Body.Close()
			log.Println(string(data))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"first proxy server not correct running"`))
			return
		}
		req, err = http.NewRequest("POST", SecondProxyAddr, strings.NewReader(`{"pid":`+strconv.FormatInt(pid, 10)+`}`))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"Internal Server Error"`))
			return
		}
		req.AddCookie(&http.Cookie{Name: "kubernetes_token", Value: tokenCookie.Value, Expires: time.Now().Add(time.Hour * 24)})
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"second proxy server not correct running"`))
			return
		}
		if res.StatusCode != http.StatusCreated {
			data, _ = ioutil.ReadAll(res.Body)
			res.Body.Close()
			log.Println(string(data))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"second proxy server not correct running"`))
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"iid":` + strconv.FormatInt(iid, 10) + `}`))
}

func updateInstance(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var (
		instance     Instance
		instanceName string
		cid          string
	)
	err := json.Unmarshal(data, &instance)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error"}`))
		return
	}
	if instance.Iid == 0 || instance.Aid == 0 || instance.Param == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"iid, aid and param are required"}`))
		return
	}
	instances, err := db.MysqlDB.Query("select instancename, cid from instance where iid=" + strconv.Itoa(instance.Iid) + " and state=0")
	defer instances.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query instance info"}`))
		return
	}
	if instances.Next() {
		err = instances.Scan(&instanceName, &cid)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query instance info"}`))
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"running instance with this iid not exists"}`))
		return
	}
	app, err := db.MysqlDB.Query("select path from app where aid=" + strconv.Itoa(instance.Aid))
	app.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query app info"}`))
		return
	}
	var path string
	if app.Next() {
		err = app.Scan(&path)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query app info"}`))
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"aid not exists in mysql"}`))
		return
	}
	cmd := exec.Command(path, "--action", "update", strconv.Itoa(claim.Uid), claim.UserName, instanceName, instance.Param)
	buf, err := cmd.Output()
	if err != nil {
		log.Println(err)
		log.Println(string(buf))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"instance update failed"}`))
		return
	}
	_, err = db.MysqlDB.Exec("update config set endtime=now() where cid=" + cid)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to update instance's config in mysql"}`))
		return
	}
	res, err := db.MysqlDB.Exec("insert config set iid=" + strconv.Itoa(instance.Iid) + ", starttime=now(), param='" + instance.Param + "'")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"instance's config creation failed"}`))
		return
	}
	new_cid, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"instance's config creation failed"}`))
		return
	}
	_, err = db.MysqlDB.Exec("update instance set cid=" + strconv.FormatInt(new_cid, 10) + " where iid=" + strconv.Itoa(instance.Iid))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"instance's update failed"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"instance update successful"}`))
}

func deleteInstance(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var (
		instance     Instance
		path         string
		instanceName string
		cid          string
		pid          string
	)
	err := json.Unmarshal(data, &instance)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error"}`))
		return
	}
	if instance.Iid == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"iid is required"}`))
		return
	}
	rows, err := db.MysqlDB.Query("select path, instancename, cid from instance, app where iid=" + strconv.Itoa(instance.Iid) + " and state=0 and instance.aid=app.aid")
	defer rows.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query instance info"}`))
		return
	}
	if rows.Next() {
		err = rows.Scan(&path, &instanceName, &cid)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query instance info"}`))
			return
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query instance info"}`))
		return
	}
	cmd := exec.Command(path, "--action", "delete", claim.UserName, instanceName)
	buf, err := cmd.Output()
	if err != nil {
		log.Println(err)
		log.Println(string(buf))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"instance deletion failed"}`))
		return
	}
	_, err = db.MysqlDB.Exec("update instance set state=1, deletetime=now() where iid=" + strconv.Itoa(instance.Iid))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to update instance in mysql"}`))
		return
	}
	_, err = db.MysqlDB.Exec("update config set endtime=now() where cid=" + cid)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to update instance's config in mysql"}`))
		return
	}
	tokenCookie, err := r.Cookie("kubernetes_token")
	proxys, err := db.MysqlDB.Query("select pid from proxy where iid=" + strconv.Itoa(instance.Iid))
	defer proxys.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query proxy info"}`))
		return
	}
	for proxys.Next() {
		err = proxys.Scan(&pid)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query proxy info"}`))
			return
		}
		_, err = db.MysqlDB.Exec("delete from proxy where pid=" + pid)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to delete proxy info"}`))
			return
		}
		req, err := http.NewRequest("DELETE", FirstProxyAddr, strings.NewReader(`{"pid":`+pid+`}`))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"Internal Server Error!"`))
			return
		}
		req.AddCookie(&http.Cookie{Name: "kubernetes_token", Value: tokenCookie.Value, Expires: time.Now().Add(time.Hour * 24)})
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"first proxy server not correct running"`))
			return
		}
		if res.StatusCode != http.StatusOK {
			data, _ = ioutil.ReadAll(res.Body)
			res.Body.Close()
			log.Println(string(data))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"first proxy server not correct running"`))
			return
		}
		req, err = http.NewRequest("DELETE", SecondProxyAddr, strings.NewReader(`{"pid":`+pid+`}`))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"Internal Server Error!"`))
			return
		}
		req.AddCookie(&http.Cookie{Name: "kubernetes_token", Value: tokenCookie.Value, Expires: time.Now().Add(time.Hour * 24)})
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"first proxy server not correct running"`))
			return
		}
		if res.StatusCode != http.StatusOK {
			data, _ = ioutil.ReadAll(res.Body)
			res.Body.Close()
			log.Println(string(data))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"message":"first proxy server not correct running"`))
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"instance delete successful"}`))
}

func InstanceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	token, err := auth.JwtAuthRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Access Unauthorized!"}`))
		return
	}
	claim := token.Claims.(*auth.KubernetesClaims)
	if r.Method == "GET" {
		queryInstance(w, r, claim)
		return
	}
	if r.Method == "POST" {
		addInstance(w, r, claim)
		return
	}
	if r.Method == "PATCH" {
		updateInstance(w, r, claim)
		return
	}
	if r.Method == "DELETE" {
		deleteInstance(w, r, claim)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"message":"only method GET, POST, PATCH and DELETE are supported"}`))
}
