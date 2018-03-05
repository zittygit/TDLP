package group

import (
	"encoding/json"
	"fmt"
	"gopkg.in/ldap.v2"
	"io/ioutil"
	"kubernetes/auth"
	"kubernetes/db"
	"kubernetes/user"
	"log"
	"net/http"
	"strconv"
)

type Group struct {
	Gid       int    `json:"gid"`
	GroupName string `json:"groupName"`
}

func queryGroup(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	var (
		gid       string
		groupName string
	)
	if claim.Role == user.ADMIN {
		r.ParseForm()
		gid = r.FormValue("gid")
		groupName = r.FormValue("groupName")
		query := "select gid, groupName from groups"
		if gid != "" {
			query += " where gid=" + gid
		} else {
			if groupName != "" {
				query += " where groupName='" + groupName + "'"
			}
		}
		groups, err := db.MysqlDB.Query(query)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query group info!"}`))
			return
		}
		defer groups.Close()
		str := `{"groups":[`
		for groups.Next() {
			err = groups.Scan(&gid, &groupName)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query group info!"}`))
				return
			} else {
				str += `{"gid":` + gid + `,"groupName":"` + groupName + `"},`
			}
		}
		if str[len(str)-1] == ',' {
			str = str[0:len(str)-1] + "]}"
		} else {
			str += `]}`
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(str))
	} else {
		groups, err := db.MysqlDB.Query("select groups.gid, groupName from groups, users where uid=" + strconv.Itoa(claim.Uid) + " and groups.gid=users.gid")
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query group info!"}`))
			return
		}
		defer groups.Close()
		if groups.Next() {
			err = groups.Scan(&gid, &groupName)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query group info!}`))
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"groups":[{"gid":` + gid + `,"groupName":"` + groupName + `"}]`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"groups":[]}`))
		}
	}
}

func createMysqlGroup(group *Group) error {
	_, err := db.MysqlDB.Exec("alter table groups auto_increment=1001")
	if err != nil {
		return err
	}
	res, err := db.MysqlDB.Exec("insert groups set groupName='" + group.GroupName + "'")
	if err != nil {
		return err
	}
	gid, err := res.LastInsertId()
	group.Gid = int(gid)
	return err
}

func createLDAPGroup(group *Group) error {
	conn, err := ldap.Dial("tcp", auth.LDAPServer+":389")
	if err != nil {
		return err
	}
	defer conn.Close()
	err = conn.Bind(auth.BindDN, auth.BindPassWord)
	if err != nil {
		return err
	}
	groupDN := fmt.Sprintf(auth.GroupDN, group.GroupName)
	request := ldap.NewAddRequest(groupDN)
	request.Attribute("objectClass", []string{"posixGroup", "top"})
	request.Attribute("userPassword", []string{""})
	request.Attribute("gidNumber", []string{strconv.Itoa(group.Gid)})
	return conn.Add(request)
}

func createGroup(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	if claim.Role != user.ADMIN {
		log.Println("only admin can add group!")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"only admin can add group!"}`))
		return
	}
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var group Group
	err := json.Unmarshal(data, &group)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error!"}`))
		return
	}
	if group.GroupName == "" {
		log.Println("groupName is required!")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"groupName is required!"}`))
		return
	}
	err = createMysqlGroup(&group)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to add group to mysql!"}`))
		return
	}
	err = createLDAPGroup(&group)
	if err != nil {
		log.Println(err)
		err = deleteMysqlGroup(&group)
		if err != nil {
			log.Println(err)
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to add group to ldap!"}`))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"gid":` + strconv.Itoa(group.Gid) + `}`))
	}
}

func deleteMysqlGroup(group *Group) error {
	_, err := db.MysqlDB.Exec("delete from groups where gid=" + strconv.Itoa(group.Gid))
	return err
}

func deleteLDAPGroup(group *Group) error {
	conn, err := ldap.Dial("tcp", auth.LDAPServer+":389")
	if err != nil {
		return err
	}
	defer conn.Close()
	err = conn.Bind(auth.BindDN, auth.BindPassWord)
	if err != nil {
		return err
	}
	groupDN := fmt.Sprintf(auth.GroupDN, group.GroupName)
	request := ldap.NewDelRequest(groupDN, nil)
	return conn.Del(request)
}

func deleteGroup(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	if claim.Role != user.ADMIN {
		log.Println("only admin can delete group!")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"only admin can delete group!"}`))
		return
	}
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var group Group
	err := json.Unmarshal(data, &group)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error!"}`))
		return
	}
	if group.Gid == 0 {
		log.Println("gid is required!")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"gid is required!"}`))
		return
	}
	groups, err := db.MysqlDB.Query("select groupName from groups where gid=" + strconv.Itoa(group.Gid))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query group's name!"}`))
		return
	}
	defer groups.Close()
	if groups.Next() {
		err = groups.Scan(&(group.GroupName))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query group's name!"}`))
			return
		}
	} else {
		log.Println("failed to query group's name using gid " + strconv.Itoa(group.Gid))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query group's name using gid ` + strconv.Itoa(group.Gid) + `"}`))
		return
	}
	err = deleteLDAPGroup(&group)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to delete group in ldap!"}`))
		return
	}
	err = deleteMysqlGroup(&group)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to delete group in mysql!"}`))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"group deleted"}`))
	}
}

func GroupHandler(w http.ResponseWriter, r *http.Request) {
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
		queryGroup(w, r, claim)
		return
	}
	if r.Method == "POST" {
		createGroup(w, r, claim)
		return
	}
	if r.Method == "DELETE" {
		deleteGroup(w, r, claim)
		return
	}
	log.Println("only method GET, POST and DELETE are supported!")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"message":"only method GET, POST and DELETE are supported!"}`))
}
