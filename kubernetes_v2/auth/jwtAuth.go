package auth

import (
	"crypto/rsa"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
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
	Role     int
}

func JwtInit(privateKeyPath, publicKeyPath string) error {
	signBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return err
	}
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return err
	}
	verifyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return err
	}
	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	return err
}

func JwtCreateToken(uid int, userName string, role int) (string, error) {
	token := jwt.New(jwt.GetSigningMethod("RS256"))
	token.Claims = &KubernetesClaims{
		&jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
		uid,
		userName,
		role,
	}
	return token.SignedString(signKey)
}

func JwtAuthRequest(r *http.Request) (*jwt.Token, error) {
	token, err := r.Cookie("kubernetes_token")
	if err == http.ErrNoCookie {
		return nil, errors.New("kubernetes token not found in cookies")
	}
	return jwt.ParseWithClaims(token.Value, &KubernetesClaims{}, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})
}
