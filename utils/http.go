package utils

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	path2 "path"
	"strings"
	"time"
)

type KeyPair struct {
	Key   string
	Value string
}

var EmptyKeyPairList = []KeyPair{}
var _HttpClient *http.Client

func init() {

	// e.g PROXY http://115.215.71.12:808
	if os.Getenv("PROXY") != "" {
		proxy, _ := url.Parse(os.Getenv("PROXY"))
		transport := &http.Transport{
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			MaxIdleConns:    10,
			IdleConnTimeout: 15 * time.Second,
		}
		_HttpClient = &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		}
	} else {
		transport := &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 15 * time.Second,
		}
		_HttpClient = &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		}
	}
}

func callHttp(methodType string, url string, requestBody string, params []KeyPair, headers []KeyPair) (string, error) {

	var body string
	var errorCatch error
	var buffer bytes.Buffer

	buffer.WriteString(url)
	if len(params) > 0 && !strings.HasSuffix(url, "?") {
		buffer.WriteString("?")
	}
	for i, param := range params {
		buffer.WriteString(param.Key)
		buffer.WriteString("=")
		buffer.WriteString(param.Value)
		if i < len(params)-1 {
			buffer.WriteString("&")
		}
	}

	req, err := http.NewRequest(methodType, buffer.String(), bytes.NewBuffer([]byte(requestBody)))
	// spew.Dump(req.Response.StatusCode)
	if err != nil {
		errorCatch = errors.New("build request failed")
	} else {
		for _, header := range headers {
			req.Header.Set(header.Key, header.Value)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			errorCatch = err
		} else {
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				errorCatch = errors.New("read resp failed")
			}
			body = string(bodyBytes)
		}
		closeBody(resp)
	}
	return body, errorCatch
}

func Post(url string, requestBody string, params []KeyPair, headers []KeyPair) (string, error) {
	return callHttp("POST", url, requestBody, params, headers)
}

func Get(url string, requestBody string, params []KeyPair, headers []KeyPair) (string, error) {
	return callHttp("GET", url, requestBody, params, headers)
}

func Delete(url string, requestBody string, params []KeyPair, headers []KeyPair) (string, error) {
	return callHttp("DELETE", url, requestBody, params, headers)
}

func closeBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
}

func JoinUrlPath(baseUrl string, path string) string {
	u, _ := url.Parse(baseUrl)
	u.Path = path2.Join(u.Path, path)
	return u.String()
}
