package auth

import (
	"encoding/json"
	"io/ioutil"
	"kubernetes/db"
	"log"
	"net/http"
	"strconv"
	"time"
)

type AuthInfo struct {
	UserName string `json:"userName"`
	PassWord string `json:"passWord"`
}

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	var (
		authInfo AuthInfo
		role     int
	)
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		log.Println("only method POST is supported!")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"message":"only method POST is supported!"}`))
		return
	}
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	err := json.Unmarshal(data, &authInfo)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error!"}`))
		return
	}
	if authInfo.UserName == "" || authInfo.PassWord == "" {
		log.Println("userName and passWord are required!")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"userName and passWord are required!"}`))
		return
	}
	entry, err := LdapAuthenticateUser(authInfo.UserName, authInfo.PassWord)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal server error!"}`))
	}
	if entry != nil {
		users, err := db.MysqlDB.Query("select role from users where userName='" + authInfo.UserName + "'")
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query role of user!"}`))
			return
		}
		defer users.Close()
		if users.Next() {
			err = users.Scan(&role)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query role of user!"}`))
			} else {
				uid, err := strconv.Atoi(entry.GetAttributeValue("uidNumber"))
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"error while converting uid!"}`))
				}
				token, err := JwtCreateToken(uid, authInfo.UserName, role)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"error while Signing Token!"}`))
				} else {
					_, err = db.MysqlDB.Exec("update users set lastLogin=now() where uid=" + entry.GetAttributeValue("uidNumber"))
					if err != nil {
						log.Println(err)
					}
					http.SetCookie(w, &http.Cookie{Name: "kubernetes_token", Value: token, Expires: time.Now().Add(time.Hour * 24)})
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"userName":"` + authInfo.UserName + `","role":` + strconv.Itoa(role) + `,"uid":"` + entry.GetAttributeValue("uidNumber") + `","gid":"` + entry.GetAttributeValue("gidNumber") + `"}`))
				}
			}
		} else {
			log.Println("user " + authInfo.UserName + " not exists in database!")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"user ` + authInfo.UserName + ` not exists in database!"}`))
		}
	} else {
		log.Println("userName or passWord is incorrect!")
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte(`{"message":"userName or passWord is incorrect!"}`))
	}
}
