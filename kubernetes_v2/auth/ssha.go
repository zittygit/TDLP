package auth

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
)

func createHash(passWord string, salt []byte) []byte {
	pass := []byte(passWord)
	str := append(pass[:], salt[:]...)
	sum := sha1.Sum(str)
	result := append(sum[:], salt[:]...)
	return result
}

func GenerateSSHA(passWord string) (string, error) {
	salt := make([]byte, 4)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	hash := createHash(passWord, salt)
	ret := "{SSHA}" + base64.StdEncoding.EncodeToString(hash)
	return ret, nil
}

func ValidateSSHA(passWord string, hash string) (bool, error) {
	if len(hash) < 7 || string(hash[0:6]) != "{SSHA}" {
		return false, nil
	}
	data, err := base64.StdEncoding.DecodeString(hash[6:])
	if len(data) < 21 || err != nil {
		return false, err
	}
	newHash := createHash(passWord, data[20:])
	hashedPW := base64.StdEncoding.EncodeToString(newHash)
	if hashedPW != hash[6:] {
		return false, nil
	}
	return true, nil
}
