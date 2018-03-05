package user

import (
	"encoding/json"
	"fmt"
	"gopkg.in/ldap.v2"
	"io/ioutil"
	"kubernetes/auth"
	"kubernetes/db"
	"kubernetes/k8s"
	"log"
	"net/http"
	"strconv"
)

type KubernetesUser struct {
	UserName string `json:"username"`
	PassWord string `json:"password"`
	Uid      int    `json:"uid"`
	Gid      int    `json:"gid"`
	Role     string `json:"role"`
	Email    string `json:"email"`
}

func queryUser(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	var (
		uid        string
		userName   string
		gid        *string
		groupName  string
		role       string
		email      string
		createTime string
	)
	if claim.Role == "admin" {
		r.ParseForm()
		gid = new(string)
		*gid = r.FormValue("gid")
		uid = r.FormValue("uid")
		userName = r.FormValue("username")
		querystring := "select uid, username, gid, role, email, createtime from user"
		if *gid != "" {
			querystring += " where gid=" + *gid
		} else {
			if uid != "" {
				querystring += " where uid=" + uid
			} else {
				if userName != "" {
					querystring += " where username='" + userName + "'"
				}
			}
		}
		users, err := db.MysqlDB.Query(querystring)
		defer users.Close()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query user info"}`))
			return
		}
		str := `{"users":[`
		for users.Next() {
			err = users.Scan(&uid, &userName, &gid, &role, &email, &createTime)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query user info"}`))
				return
			}
			if gid == nil {
				gid = new(string)
				*gid = "0"
				groupName = ""
			} else {
				group, err := db.MysqlDB.Query("select groupname from usergroup where gid=" + *gid)
				group.Close()
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"failed to query group info"}`))
					return
				}
				if group.Next() {
					err = group.Scan(&groupName)
					if err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(`{"message":"failed to query group info"}`))
						return
					}
				} else {
					groupName = ""
				}
			}
			str += `{"uid":` + uid + `,"username":"` + userName + `","gid":` + *gid + `,"groupname":"` + groupName + `","role":"` + role + `","email":"` + email + `","createtime":"` + createTime + `"},`
		}
		if str[len(str)-1] == ',' {
			str = str[0:len(str)-1] + "]}"
		} else {
			str += `]}`
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(str))
	} else {
		user, err := db.MysqlDB.Query("select uid, username, gid, role, email, createtime from user")
		user.Close()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query user info"}`))
			return
		}
		if user.Next() {
			err = user.Scan(&uid, &userName, &gid, &role, &email, &createTime)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query user info"}`))
				return
			}
			if gid == nil {
				gid = new(string)
				*gid = "0"
				groupName = ""
			} else {
				group, err := db.MysqlDB.Query("select groupname from usergroup where gid=" + *gid)
				group.Close()
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"failed to query group info"}`))
					return
				}
				if group.Next() {
					err = group.Scan(&groupName)
					if err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(`{"message":"failed to query group info"}`))
						return
					}
				} else {
					groupName = ""
				}
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"uid":` + uid + `,"username":"` + userName + `","gid":` + *gid + `,"groupname":"` + groupName + `","role":"` + role + `","email":"` + email + `","createtime":"` + createTime + `"}`))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query user info"}`))
		}
	}
}

func addMysqlUser(user *KubernetesUser) error {
	_, err := db.MysqlDB.Exec("alter table user auto_increment=1001")
	if err != nil {
		return err
	}
	queryString := "insert user set username='" + user.UserName + "', role='" + user.Role + "', email='" + user.Email + "', createtime=now()"
	if user.Gid != 0 {
		queryString += ", gid=" + strconv.Itoa(user.Gid)
	}
	res, err := db.MysqlDB.Exec(queryString)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	user.Uid = int(id)
	return err
}

func addLdapUser(user *KubernetesUser) error {
	conn, err := ldap.Dial("tcp", auth.LdapServer+":389")
	if err != nil {
		return err
	}
	defer conn.Close()
	err = conn.Bind(auth.BindDN, auth.BindPassword)
	if err != nil {
		return err
	}
	userDN := fmt.Sprintf(auth.UserDN, user.UserName)
	request := ldap.NewAddRequest(userDN)
	request.Attribute("uid", []string{user.UserName})
	request.Attribute("cn", []string{user.UserName})
	request.Attribute("sn", []string{user.UserName})
	request.Attribute("mail", []string{user.Email})
	request.Attribute("objectClass", []string{"person", "organizationalPerson", "inetOrgPerson", "posixAccount", "top", "shadowAccount"})
	request.Attribute("shadowLastChange", []string{"0"})
	request.Attribute("shadowMin", []string{"0"})
	request.Attribute("shadowMax", []string{"99999"})
	request.Attribute("shadowWarning", []string{"0"})
	request.Attribute("loginShell", []string{"/bin/bash"})
	request.Attribute("uidNumber", []string{strconv.Itoa(user.Uid)})
	request.Attribute("gidNumber", []string{strconv.Itoa(user.Gid)})
	request.Attribute("homeDirectory", []string{"/home/" + user.UserName})
	request.Attribute("userPassword", []string{user.PassWord})
	return conn.Add(request)
}

func addUser(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	if claim.Role != "admin" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"only admin can use this interface!"}`))
		return
	}
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var user KubernetesUser
	err := json.Unmarshal(data, &user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error"}`))
		return
	}
	if user.UserName == "" || user.PassWord == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"username and password are required"}`))
		return
	}
	user.PassWord, err = auth.GenerateSSHA(user.PassWord)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"encrypt password failed"}`))
		return
	}
	err = k8s.CreateNameSpace(user.UserName)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to create kubernetes namespace for user"}`))
		return
	}
	err = addMysqlUser(&user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to add user to mysql"}`))
		err = k8s.DeleteNameSpace(user.UserName)
		if err != nil {
			log.Println(err)
		}
		return
	}
	err = addLdapUser(&user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to add user to ldap"}`))
		err = k8s.DeleteNameSpace("test")
		if err != nil {
			log.Println(err)
		}
		err = deleteMysqlUser(&user)
		if err != nil {
			log.Println(err)
		}
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"uid":` + strconv.Itoa(user.Uid) + `}`))
	}
}

func deleteMysqlUser(user *KubernetesUser) error {
	_, err := db.MysqlDB.Exec("delete from user where uid=" + strconv.Itoa(user.Uid))
	return err
}

func deleteLdapUser(user *KubernetesUser) error {
	conn, err := ldap.Dial("tcp", auth.LdapServer+":389")
	if err != nil {
		return err
	}
	defer conn.Close()
	err = conn.Bind(auth.BindDN, auth.BindPassword)
	if err != nil {
		return err
	}
	userDN := fmt.Sprintf(auth.UserDN, user.UserName)
	request := ldap.NewDelRequest(userDN, nil)
	return conn.Del(request)
}

func deleteUser(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	if claim.Role != "admin" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"only admin can use this interface!"}`))
		return
	}
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var user KubernetesUser
	err := json.Unmarshal(data, &user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error"}`))
		return
	}
	if user.Uid == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"uid is required"}`))
		return
	}
	users, err := db.MysqlDB.Query("select username from user where uid=" + strconv.Itoa(user.Uid))
	defer users.Close()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query user's username"}`))
		return
	}
	if users.Next() {
		err = users.Scan(&user.UserName)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query user's username"}`))
			return
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query user's username"}`))
		return
	}
	err = k8s.DeleteNameSpace(user.UserName)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to delete user's namespace in kubernetes"}`))
		return
	}
	err = deleteLdapUser(&user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to delete user in ldap"}`))
		return
	}
	err = deleteMysqlUser(&user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to delete user in mysql"}`))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"user deleted"}`))
	}
}

func updateMysqlUser(user *KubernetesUser) error {
	change := false
	queryString := "update user set"
	if user.Gid != 0 {
		queryString += " gid=" + strconv.Itoa(user.Gid)
		change = true
	}
	if user.Role != "" {
		if change {
			queryString += ", role='" + user.Role + "'"
		} else {
			queryString += " role='" + user.Role + "'"
			change = true
		}
	}
	if user.Email != "" {
		if change {
			queryString += ", email='" + user.Email + "'"
		} else {
			queryString += " email='" + user.Email + "'"
			change = true
		}
	}
	if change {
		queryString += " where uid=" + strconv.Itoa(user.Uid)
		_, err := db.MysqlDB.Exec(queryString)
		return err
	} else {
		return nil
	}
}

func updateLdapUser(user *KubernetesUser) error {
	conn, err := ldap.Dial("tcp", auth.LdapServer+":389")
	if err != nil {
		return err
	}
	defer conn.Close()
	err = conn.Bind(auth.BindDN, auth.BindPassword)
	if err != nil {
		return err
	}
	userDN := fmt.Sprintf(auth.UserDN, user.UserName)
	request := ldap.NewModifyRequest(userDN)
	change := false
	if user.PassWord != "" {
		user.PassWord, err = auth.GenerateSSHA(user.PassWord)
		if err != nil {
			return err
		}
		request.Replace("userPassword", []string{user.PassWord})
		change = true
	}
	if user.Email != "" {
		request.Replace("mail", []string{user.Email})
		change = true
	}
	if user.Gid != 0 {
		request.Replace("gidNumber", []string{strconv.Itoa(user.Gid)})
		change = true
	}
	if change {
		err = conn.Modify(request)
		return err
	} else {
		return nil
	}
}

func updateUser(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var user KubernetesUser
	err := json.Unmarshal(data, &user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error"}`))
		return
	}
	if claim.Role == "admin" {
		if user.Uid == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"uid is required"}`))
			return
		}
		users, err := db.MysqlDB.Query("select username from user where uid=" + strconv.Itoa(user.Uid))
		defer users.Close()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query user's username"}`))
			return
		}
		if users.Next() {
			err = users.Scan(&user.UserName)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query user's username"}`))
				return
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query user's username"}`))
			return
		}
	} else {
		user.Uid = claim.Uid
		user.UserName = claim.UserName
		user.Role = ""
		user.Gid = 0
	}
	err = updateLdapUser(&user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to update user in ldap"}`))
		return
	}
	err = updateMysqlUser(&user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to update user in mysql"}`))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"user updated"}`))
	}
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
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
		queryUser(w, r, claim)
		return
	}
	if r.Method == "POST" {
		addUser(w, r, claim)
		return
	}
	if r.Method == "DELETE" {
		deleteUser(w, r, claim)
		return
	}
	if r.Method == "PATCH" {
		updateUser(w, r, claim)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"message":"only method GET, POST, DELETE and PATCH are supported"}`))
}
