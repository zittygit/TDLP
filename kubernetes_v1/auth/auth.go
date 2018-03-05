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
	UserName string `json:"username"`
	PassWord string `json:"password"`
}

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"message":"only method POST is supported"}`))
		return
	}
	data, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	var authInfo AuthInfo
	err := json.Unmarshal(data, &authInfo)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"post data format error"}`))
		return
	}
	if authInfo.UserName == "" || authInfo.PassWord == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"username and password are required"}`))
		return
	}
	entry := LdapAuthenticateUser(authInfo.UserName, authInfo.PassWord)
	if entry != nil {
		rows, err := db.MysqlDB.Query("select role from user where username='" + authInfo.UserName + "'")
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"failed to query role for user"}`))
			return
		}
		if rows.Next() {
			var role string
			err = rows.Scan(&role)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"failed to query role for user"}`))
			} else {
				uid, _ := strconv.Atoi(entry.GetAttributeValue("uidNumber"))
				token, err := JwtCreateToken(uid, authInfo.UserName, role)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"error while Signing Token"}`))
				} else {
					http.SetCookie(w, &http.Cookie{Name: "kubernetes_token", Value: token, Expires: time.Now().Add(time.Hour * 24)})
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"username":"` + authInfo.UserName + `","role":"` + role + `","uid":"` + entry.GetAttributeValue("uidNumber") + `","gid":"` + entry.GetAttributeValue("gidNumber") + `","homedir":"` + entry.GetAttributeValue("homeDirectory") + `"}`))
				}
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"user not exists in database"}`))
		}
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte(`{"message":"username or password is incorrect"}`))
	}
}
