package template

import (
	"encoding/json"
	"io/ioutil"
	"kubernetes/auth"
	"kubernetes/db"
	"kubernetes/user"
	"log"
	"net/http"
	"strconv"
)

type Template struct {
	Tid          int    `json:"tid"`
	TemplateName string `json:"templateName"`
	Path         string `json:"path"`
	Info         string `json:"info"`
	Param        string `json:"param"`
}

func queryTemplate(w http.ResponseWriter, r *http.Request) {
	var (
		tid          string
		templateName string
		path         string
		info         string
		param        string
	)
	r.ParseForm()
	tid = r.FormValue("tid")
	templateName = r.FormValue("templateName")
	query := "select tid, templateName, path, info, param from templates"
	if tid != "" {
		query += " where tid=" + tid
	} else {
		if templateName != "" {
			query += " where templateName='" + templateName + "'"
		}
	}
	templates, err := db.MysqlDB.Query(query)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query template info!"}`))
		return
	}
	defer templates.Close()
	str := `{"templates":[`
	for templates.Next() {
		err = templates.Scan(&tid, &templateName, &path, &info, &param)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query template info!"}`))
			return
		}
		str += `{"tid":` + tid + `,"templateName":"` + templateName + `","path":"` + path + `","info":"` + info + `","param":` + strconv.Quote(param) + `},`
	}
	if str[len(str)-1] == ',' {
		str = str[0:len(str)-1] + "]}"
	} else {
		str += `]}`
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(str))
}

func createTemplate(w http.ResponseWriter, r *http.Request) {
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var template Template
	err := json.Unmarshal(data, &template)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error!"}`))
		return
	}
	if template.TemplateName == "" || template.Path == "" || template.Info == "" || template.Param == "" {
		log.Println("templateName, path, info and param are required!")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"templateName, path, info and param are required!"}`))
		return
	}
	insert := "insert template set templateName='" + template.TemplateName + "', path='" + template.Path + "', info='" + template.Info + "', param='" + template.Param + "'"
	res, err := db.MysqlDB.Exec(insert)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to add template to mysql!"}`))
	} else {
		tid, err := res.LastInsertId()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"failed to add template to mysql!"}`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"tid":` + strconv.FormatInt(tid, 10) + `}`))
		}
	}
}

func updateTemplate(w http.ResponseWriter, r *http.Request) {
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var template Template
	err := json.Unmarshal(data, &template)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error!"}`))
		return
	}
	if template.Tid == 0 {
		log.Println("tid is required!")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"tid is required!"}`))
		return
	}
	if template.Path != "" {
		log.Println("can not change template's path because app will use it!")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"can not change template's path because app will use it!"}`))
		return
	}
	change := false
	update := "update templates set"
	if template.TemplateName != "" {
		update += " name='" + template.TemplateName + "'"
		change = true
	}
	if template.Info != "" {
		if change {
			update += ", info='" + template.Info + "'"
		} else {
			update += " info='" + template.Info + "'"
			change = true
		}
	}
	if template.Param != "" {
		if change {
			update += ", param='" + template.Param + "'"
		} else {
			update += " param='" + template.Param + "'"
			change = true
		}
	}
	update += " where tid=" + strconv.Itoa(template.Tid)
	if change {
		_, err := db.MysqlDB.Exec(update)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"failed to update template in mysql!"}`))
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"template updated!"}`))
}

func deleteTemplate(w http.ResponseWriter, r *http.Request) {
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var template Template
	err := json.Unmarshal(data, &template)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error!"}`))
		return
	}
	if template.Tid == 0 {
		log.Println("tid is required!")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"tid is required!"}`))
		return
	}
	apps, err := db.MysqlDB.Query("select aid from apps where state=0 and tid=" + strconv.Itoa(template.Tid))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query app info!"}`))
		return
	}
	defer apps.Close()
	if apps.Next() {
		log.Println("running instance of template with tid " + strconv.Itoa(template.Tid) + " exists, can not delete this template!")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"running instance of template with tid ` + strconv.Itoa(template.Tid) + ` exists, can not delete this template!"}`))
		return
	}
	_, err = db.MysqlDB.Exec("delete from templates where tid=" + strconv.Itoa(template.Tid))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to delete template in mysql!"}`))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"template deleted!"}`))
	}
}

func TemplateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	token, err := auth.JwtAuthRequest(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Access Unauthorized!"}`))
		return
	}
	if r.Method == "GET" {
		queryTemplate(w, r)
		return
	}
	claim := token.Claims.(*auth.KubernetesClaims)
	if claim.Role != user.ADMIN {
		log.Println("user " + claim.UserName + "intend to manage template!")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"only admin can use this interface!"}`))
		return
	}
	if r.Method == "POST" {
		createTemplate(w, r)
		return
	}
	if r.Method == "PATCH" {
		updateTemplate(w, r)
		return
	}
	if r.Method == "DELETE" {
		deleteTemplate(w, r)
		return
	}
	log.Println("only method GET, POST, PATCH and DELETE are supported")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"message":"only method GET, POST, PATCH and DELETE are supported"}`))
}
