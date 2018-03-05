package user

import (
	"encoding/json"
	"errors"
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

const (
	USER     = 0
	ADMIN    = 3
	INACTIVE = 0
	ACTIVE   = 1
)

type User struct {
	UserName     string `json:"userName"`
	PassWord     string `json:"PassWord"`
	Uid          int    `json:"uid"`
	Gid          int    `json:"gid"`
	Role         int    `json:"role"`
	Email        string `json:"email"`
	RealName     string `json:"realName"`
	Phone        string `json:"phone"`
	Organization string `json:"organization"`
	Avatar       string `json:"avatar"`
}

func IsActive(claim *auth.KubernetesClaims) (int, error) {
	var active int
	users, err := db.MysqlDB.Query("select active from users where uid=" + strconv.Itoa(claim.Uid))
	if err != nil {
		return INACTIVE, err
	}
	defer users.Close()
	if users.Next() {
		err = users.Scan(&active)
		return active, err
	} else {
		return INACTIVE, errors.New("can not found user info!")
	}
}

func CheckActive(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) int {
	active, err := IsActive(claim)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal server error!"}`))
		return INACTIVE
	}
	if active == INACTIVE {
		log.Println("user " + claim.UserName + " is not a active user!")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"user ` + claim.UserName + ` is not a active user!"}`))
	}
	return active
}

func queryUser(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	var (
		uid          string
		userName     string
		gid          *string
		groupName    *string
		realName     string
		phone        string
		organization string
		avatar       string
		role         int
		email        string
		createTime   string
		lastLogin    string
		cond         string
	)
	if claim.Role == ADMIN {
		r.ParseForm()
		gid = new(string)
		*gid = r.FormValue("gid")
		uid = r.FormValue("uid")
		userName = r.FormValue("userName")
		if *gid != "" {
			cond = " where gid=" + *gid
		} else {
			if uid != "" {
				cond = " where uid=" + uid
			} else {
				if userName != "" {
					cond = " where userName='" + userName + "'"
				}
			}
		}
		query := "select uid, userName, temp.gid, groupName, realName, phone, organization, avatar, role, email, createTime, lastLogin from (select uid, userName, gid, realName, phone, organization, avatar, role, email, createTime, lastLogin from users" + cond + ") temp left join groups on temp.gid=groups.gid"
		users, err := db.MysqlDB.Query(query)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query user info!"}`))
			return
		}
		defer users.Close()
		str := `{"users":[`
		for users.Next() {
			err = users.Scan(&uid, &userName, &gid, &groupName, &realName, &phone, &organization, &avatar, &role, &email, &createTime, &lastLogin)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query user info!"}`))
				return
			}
			if gid == nil {
				gid = new(string)
				*gid = "0"
			}
			if groupName == nil {
				groupName = new(string)
				*groupName = ""
			}
			str += `{"uid":` + uid + `,"userName":"` + userName + `","gid":` + *gid + `,"groupName":"` + *groupName + `","realName":"` + realName + `","phone":"` + phone + `","organization":"` + organization + `","avatar":"` + avatar + `","role":"` + strconv.Itoa(role) + `","email":"` + email + `","createtime":"` + createTime + `","lastLogin":"` + lastLogin + `"},`
		}
		if str[len(str)-1] == ',' {
			str = str[0:len(str)-1] + "]}"
		} else {
			str += `]}`
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(str))
	} else {
		users, err := db.MysqlDB.Query("select uid, userName, temp.gid, groupName, realName, phone, organization, avatar, role, email, createtime, lastLogin from (select uid, userName, gid, realName, phone, organization, avatar, role, email, createTime, lastLogin from users where uid=" + strconv.Itoa(claim.Uid) + ") temp left join groups on temp.gid=groups.gid")
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query user info!"}`))
			return
		}
		defer users.Close()
		if users.Next() {
			err = users.Scan(&uid, &userName, &gid, &groupName, &realName, &phone, &organization, &avatar, &role, &email, &createTime, &lastLogin)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query user info!"}`))
				return
			}
			if gid == nil {
				gid = new(string)
				*gid = "0"
			}
			if groupName == nil {
				groupName = new(string)
				*groupName = ""
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"uid":` + uid + `,"userName":"` + userName + `","gid":` + *gid + `,"groupName":"` + *groupName + `","realName":"` + realName + `","phone":"` + phone + `","organization":"` + organization + `","avatar":"` + avatar + `","role":"` + strconv.Itoa(role) + `","email":"` + email + `","createtime":"` + createTime + `","lastLogin":"` + lastLogin + `"}`))
		} else {
			log.Println("user with uid " + strconv.Itoa(claim.Uid) + "not found!")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"user with uid " + strconv.Itoa(claim.Uid) + "not found!"}`))
		}
	}
}

func createMysqlUser(user *User) error {
	_, err := db.MysqlDB.Exec("alter table users auto_increment=1001")
	if err != nil {
		return err
	}
	insert := "insert users set userName='" + user.UserName + "', realName='" + user.RealName + "', phone='" + user.Phone + "', organization='" + user.Organization + "', avatar='" + user.Avatar + "', role='" + strconv.Itoa(user.Role) + "', email='" + user.Email + "', createTime=now(), lastLogin=now()"
	if user.Gid != 0 {
		insert += ", gid=" + strconv.Itoa(user.Gid)
	}
	res, err := db.MysqlDB.Exec(insert)
	if err != nil {
		return err
	}
	uid, err := res.LastInsertId()
	user.Uid = int(uid)
	return err
}

func createLDAPUser(user *User) error {
	conn, err := ldap.Dial("tcp", auth.LDAPServer+":389")
	if err != nil {
		return err
	}
	defer conn.Close()
	err = conn.Bind(auth.BindDN, auth.BindPassWord)
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

func createUser(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	if claim.Role != ADMIN {
		log.Println("only admin can use this interface!")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"only admin can use this interface!"}`))
		return
	}
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var user User
	err := json.Unmarshal(data, &user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error!"}`))
		return
	}
	if user.UserName == "" || user.PassWord == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"userName and passWord are required!"}`))
		return
	}
	user.PassWord, err = auth.GenerateSSHA(user.PassWord)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"encrypt password failed!"}`))
		return
	}
	err = k8s.CreateNameSpace(user.UserName)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to create kubernetes namespace for user!"}`))
		return
	}
	err = createMysqlUser(&user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to add user to mysql!"}`))
		err = k8s.DeleteNameSpace(user.UserName)
		if err != nil {
			log.Println(err)
		}
		return
	}
	err = createLDAPUser(&user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to add user to ldap!"}`))
		err = k8s.DeleteNameSpace(user.UserName)
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

func deleteMysqlUser(user *User) error {
	_, err := db.MysqlDB.Exec("delete from users where uid=" + strconv.Itoa(user.Uid))
	return err
}

func deleteLDAPUser(user *User) error {
	conn, err := ldap.Dial("tcp", auth.LDAPServer+":389")
	if err != nil {
		return err
	}
	defer conn.Close()
	err = conn.Bind(auth.BindDN, auth.BindPassWord)
	if err != nil {
		return err
	}
	userDN := fmt.Sprintf(auth.UserDN, user.UserName)
	request := ldap.NewDelRequest(userDN, nil)
	return conn.Del(request)
}

func deleteUser(w http.ResponseWriter, r *http.Request, claim *auth.KubernetesClaims) {
	if claim.Role != ADMIN {
		log.Println("only admin can use this interface!")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"only admin can use this interface!"}`))
		return
	}
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var user User
	err := json.Unmarshal(data, &user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error!"}`))
		return
	}
	if user.Uid == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"uid is required!"}`))
		return
	}
	users, err := db.MysqlDB.Query("select userName from users where uid=" + strconv.Itoa(user.Uid))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query user's username!"}`))
		return
	}
	defer users.Close()
	if users.Next() {
		err = users.Scan(&user.UserName)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query user's username!"}`))
			return
		}
	} else {
		log.Println("failed to query user's username!")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"failed to query user's username!"}`))
		return
	}
	err = k8s.DeleteNameSpace(user.UserName)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to delete user's namespace in kubernetes!"}`))
		return
	}
	err = deleteLDAPUser(&user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to delete user in ldap!"}`))
		return
	}
	err = deleteMysqlUser(&user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to delete user in mysql!"}`))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"user deleted"}`))
	}
}

func updateMysqlUser(user *User) error {
	change := false
	update := "update users set"
	if user.Gid != 0 {
		update += " gid=" + strconv.Itoa(user.Gid)
		change = true
	}
	if user.Role != 0 {
		if change {
			update += ", role='" + strconv.Itoa(user.Role) + "'"
		} else {
			update += " role='" + strconv.Itoa(user.Role) + "'"
			change = true
		}
	}
	if user.RealName != "" {
		if change {
			update += ", realName='" + user.RealName + "'"
		} else {
			update += " realName='" + user.RealName + "'"
			change = true
		}
	}
	if user.Phone != "" {
		if change {
			update += ", phone='" + user.Phone + "'"
		} else {
			update += " phone='" + user.Phone + "'"
			change = true
		}
	}
	if user.Organization != "" {
		if change {
			update += ", organization='" + user.Organization + "'"
		} else {
			update += " organization='" + user.Organization + "'"
			change = true
		}
	}
	if user.Avatar != "" {
		if change {
			update += ", avatar='" + user.Avatar + "'"
		} else {
			update += " avatar='" + user.Avatar + "'"
			change = true
		}
	}
	if user.Email != "" {
		if change {
			update += ", email='" + user.Email + "'"
		} else {
			update += " email='" + user.Email + "'"
			change = true
		}
	}
	if change {
		update += " where uid=" + strconv.Itoa(user.Uid)
		_, err := db.MysqlDB.Exec(update)
		return err
	} else {
		return nil
	}
}

func updateLDAPUser(user *User) error {
	conn, err := ldap.Dial("tcp", auth.LDAPServer+":389")
	if err != nil {
		return err
	}
	defer conn.Close()
	err = conn.Bind(auth.BindDN, auth.BindPassWord)
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
	var user User
	err := json.Unmarshal(data, &user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error!"}`))
		return
	}
	if claim.Role == ADMIN {
		if user.Uid == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"uid is required!"}`))
			return
		}
		users, err := db.MysqlDB.Query("select userName from users where uid=" + strconv.Itoa(user.Uid))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query user's username!"}`))
			return
		}
		defer users.Close()
		if users.Next() {
			err = users.Scan(&user.UserName)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query user's username!"}`))
				return
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query user's username!"}`))
			return
		}
	} else {
		user.Uid = claim.Uid
		user.UserName = claim.UserName
		user.Role = 0
		user.Gid = 0
	}
	err = updateLDAPUser(&user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to update user in ldap!"}`))
		return
	}
	err = updateMysqlUser(&user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"failed to update user in mysql!"}`))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"user updated!"}`))
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
		createUser(w, r, claim)
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
	log.Println("only method GET, POST, DELETE and PATCH are supported!")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"message":"only method GET, POST, DELETE and PATCH are supported!"}`))
}
