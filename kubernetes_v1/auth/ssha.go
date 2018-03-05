package auth

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
)

func createHash(password string, salt []byte) []byte {
	pass := []byte(password)
	str := append(pass[:], salt[:]...)
	sum := sha1.Sum(str)
	result := append(sum[:], salt[:]...)
	return result
}

func GenerateSSHA(password string) (string, error) {
	salt := make([]byte, 4)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	hash := createHash(password, salt)
	ret := "{SSHA}" + base64.StdEncoding.EncodeToString(hash)
	return ret, nil
}

func ValidateSSHA(password string, hash string) bool {
	if len(hash) < 7 || string(hash[0:6]) != "{SSHA}" {
		return false
	}
	data, err := base64.StdEncoding.DecodeString(hash[6:])
	if len(data) < 21 || err != nil {
		return false
	}
	newhash := createHash(password, data[20:])
	hashedpw := base64.StdEncoding.EncodeToString(newhash)
	if hashedpw != hash[6:] {
		return false
	}
	return true
}
