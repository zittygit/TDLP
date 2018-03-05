package k8s

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	BearerToken  string
	K8sApiServer string
)

func CreateNameSpace(username string) error {
	req, err := http.NewRequest("POST", K8sApiServer+"/namespaces", strings.NewReader(`{"apiVersion":"v1","metadata":{"name":"`+username+`"}}`))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", BearerToken)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == 201 {
		return nil
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	return errors.New(string(data))
}

func DeleteNameSpace(username string) error {
	req, err := http.NewRequest("DELETE", K8sApiServer+"/namespaces/"+username, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", BearerToken)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode == 200 {
		return nil
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	return errors.New(string(data))
}
