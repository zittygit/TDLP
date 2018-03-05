package user

import (
	"encoding/json"
	"fmt"
	"gopkg.in/ldap.v2"
	"io/ioutil"
	"kubernetes/auth"
	"kubernetes/db"
	"log"
	"net/http"
	"strconv"
)

type KubernetesGroup struct {
	Gid       int    `json:"gid"`
	GroupName string `json:"groupname"`
}

func queryGroup(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	var (
		gid       *string
		groupName string
	)
	if claim.Role == "admin" {
		r.ParseForm()
		gid = new(string)
		*gid = r.FormValue("gid")
		groupName = r.FormValue("groupname")
		querystring := "select gid, groupname from usergroup"
		if *gid != "" {
			querystring += " where gid=" + *gid
		} else {
			if groupName != "" {
				querystring += " where groupname='" + groupName + "'"
			}
		}
		groups, err := db.MysqlDB.Query(querystring)
		defer groups.Close()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query group info"}`))
			return
		}
		str := `{"groups":[`
		for groups.Next() {
			err = groups.Scan(&gid, &groupName)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query group info"}`))
				return
			} else {
				str += `{"gid":` + *gid + `,"groupname":"` + groupName + `"},`
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
		users, err := db.MysqlDB.Query("select gid from user where uid=" + strconv.Itoa(claim.Uid))
		defer users.Close()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query user info"}`))
			return
		}
		if users.Next() {
			err = users.Scan(&gid)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query user info"}`))
				return
			}
			if gid == nil {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"groups":[]}`))
				return
			}
			groups, err := db.MysqlDB.Query("select groupname from usergroup where gid=" + *gid)
			defer groups.Close()
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query group info"}`))
				return
			}
			if groups.Next() {
				err = groups.Scan(&groupName)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"failed to query group info"}`))
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"groups":[{"gid":` + *gid + `,"groupname":"` + groupName + `"}]`))
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"groups":[]}`))
	}
}

func addMysqlGroup(group *KubernetesGroup) error {
	_, err := db.MysqlDB.Exec("alter table usergroup auto_increment=1001")
	if err != nil {
		return err
	}
	res, err := db.MysqlDB.Exec("insert usergroup set groupname='" + group.GroupName + "'")
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	group.Gid = int(id)
	return err
}

func addLdapGroup(group *KubernetesGroup) error {
	conn, err := ldap.Dial("tcp", auth.LdapServer+":389")
	if err != nil {
		return err
	}
	defer conn.Close()
	err = conn.Bind(auth.BindDN, auth.BindPassword)
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

func addGroup(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	if claim.Role != "admin" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"only admin can use this interface!"}`))
		return
	}
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var group KubernetesGroup
	err := json.Unmarshal(data, &group)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error"}`))
		return
	}
	if group.GroupName == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"groupname is required"}`))
		return
	}
	err = addMysqlGroup(&group)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to add group to mysql"}`))
		return
	}
	err = addLdapGroup(&group)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to add group to ldap"}`))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"gid":` + strconv.Itoa(group.Gid) + `}`))
	}
}

func deleteMysqlGroup(group *KubernetesGroup) error {
	_, err := db.MysqlDB.Exec("delete from usergroup where gid=" + strconv.Itoa(group.Gid))
	return err
}

func deleteLdapGroup(group *KubernetesGroup) error {
	conn, err := ldap.Dial("tcp", auth.LdapServer+":389")
	if err != nil {
		return err
	}
	defer conn.Close()
	err = conn.Bind(auth.BindDN, auth.BindPassword)
	if err != nil {
		return err
	}
	groupDN := fmt.Sprintf(auth.GroupDN, group.GroupName)
	request := ldap.NewDelRequest(groupDN, nil)
	return conn.Del(request)
}

func deleteGroup(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	if claim.Role != "admin" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"only admin can use this interface!"}`))
		return
	}
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var group KubernetesGroup
	err := json.Unmarshal(data, &group)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error"}`))
		return
	}
	if group.Gid == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"gid is required"}`))
		return
	}
	groups, err := db.MysqlDB.Query("select groupname from usergroup where gid=" + strconv.Itoa(group.Gid))
	defer groups.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query group's groupname"}`))
		return
	}
	if groups.Next() {
		err = groups.Scan(&(group.GroupName))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query group's groupname"}`))
			return
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query group's groupname"}`))
		return
	}
	err = deleteLdapGroup(&group)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to delete group in ldap"}`))
		return
	}
	err = deleteMysqlGroup(&group)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to delete group in mysql"}`))
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
		addGroup(w, r, claim)
		return
	}
	if r.Method == "DELETE" {
		deleteGroup(w, r, claim)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"message":"only method GET, POST and DELETE are supported"}`))
}
