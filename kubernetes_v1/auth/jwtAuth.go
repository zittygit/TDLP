package auth

import (
	"crypto/rsa"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

type KubernetesClaims struct {
	*jwt.StandardClaims
	Uid      int
	UserName string
	Role     string
}

func JwtInit(privateKeyPath string, publicKeyPath string) {
	signBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalln(err)
	}
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatalln(err)
	}
	verifyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		log.Fatalln(err)
	}
	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		log.Fatalln(err)
	}
}

func JwtCreateToken(uid int, username string, role string) (string, error) {
	token := jwt.New(jwt.GetSigningMethod("RS256"))
	token.Claims = &KubernetesClaims{
		&jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
		uid,
		username,
		role,
	}
	return token.SignedString(signKey)
}

func JwtAuthRequest(r *http.Request) (*jwt.Token, error) {
	token, err := r.Cookie("kubernetes_token")
	if err == http.ErrNoCookie {
		return nil, errors.New("token cookie not found")
	}
	return jwt.ParseWithClaims(token.Value, &KubernetesClaims{}, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})
}
