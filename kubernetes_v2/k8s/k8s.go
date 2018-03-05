package k8s

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"kubernetes/conf"
	"net/http"
	"strings"
)

var (
	LDAPServer1  string
	LDAPServer2  string
	bearerToken  string
	k8sAPIServer string
	k8sClient    *http.Client
)

func InitK8s() error {
	config := new(conf.Config)
	err := config.InitConfig("k8s/k8s.ini")
	if err != nil {
		return err
	}
	LDAPServer1 = config.Get("LDAPServer1")
	if LDAPServer1 == "" {
		return errors.New("LDAPServer1 must be set")
	}
	LDAPServer2 = config.Get("LDAPServer2")
	if LDAPServer2 == "" {
		return errors.New("LDAPServer2 must be set")
	}
	bearerToken = config.Get("bearerToken")
	if bearerToken == "" {
		return errors.New("bearerToken must be set")
	}
	k8sAPIServer = config.Get("k8sAPIServer")
	if k8sAPIServer == "" {
		return errors.New("k8sAPIServer must be set")
	}
	rootCA := config.Get("rootCA")
	if rootCA == "" {
		return errors.New("rootCA must be set")
	}
	pool := x509.NewCertPool()
	caPem, err := ioutil.ReadFile(rootCA)
	if err != nil {
		return err
	}
	pool.AppendCertsFromPEM(caPem)
	k8sClient = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}, DisableKeepAlives: true}}
	return nil
}

func sendRequest(method, url string, template io.Reader, statusCode int) (string, error) {
	req, err := http.NewRequest(method, url, template)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", bearerToken)
	res, err := k8sClient.Do(req)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", err
	}
	if res.StatusCode == statusCode {
		return string(data), nil
	}
	return "", errors.New(string(data))
}

func CreateNameSpace(userName string) error {
	_, err := sendRequest("POST", k8sAPIServer+"/namespaces", strings.NewReader(`{"apiVersion":"v1","metadata":{"name":"`+userName+`"}}`), http.StatusCreated)
	if err != nil {
		return err
	}
	return nil
}

func DeleteNameSpace(userName string) error {
	_, err := sendRequest("DELETE", k8sAPIServer+"/namespaces/"+userName, nil, http.StatusOK)
	if err != nil {
		return err
	}
	return nil
}

func CreateReplicationController(userName, template string) error {
	_, err := sendRequest("POST", k8sAPIServer+"/namespaces/"+userName+"/replicationcontrollers", strings.NewReader(template), http.StatusCreated)
	if err != nil {
		return err
	}
	return nil
}

func DeleteReplicationController(userName, name, template string) error {
	_, err := sendRequest("PUT", k8sAPIServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, strings.NewReader(template), http.StatusOK)
	if err != nil {
		return err
	}
	_, err = sendRequest("DELETE", k8sAPIServer+"/namespaces/"+userName+"/replicationcontrollers/"+name, nil, http.StatusOK)
	if err != nil {
		return err
	}
	return nil
}

func CreateService(userName, template string) (string, error) {
	data, err := sendRequest("POST", k8sAPIServer+"/namespaces/"+userName+"/services", strings.NewReader(template), http.StatusCreated)
	if err != nil {
		return "", err
	} else {
		return gjson.Get(data, "spec.clusterIP").String(), nil
	}
}

func DeleteService(userName, name string) error {
	_, err := sendRequest("DELETE", k8sAPIServer+"/namespaces/"+userName+"/services/"+name, nil, http.StatusOK)
	if err != nil {
		return err
	}
	return nil
}
