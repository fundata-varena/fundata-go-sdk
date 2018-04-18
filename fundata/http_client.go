package fundata

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Header struct {
	AcceptApiKey   string `json:"Accept-ApiKey"`
	AcceptApiNonce string `json:"Accept-ApiNonce"`
	AcceptApiTime  string `json:"Accept-ApiTime"`
	AcceptApiSign  string `json:"Accept-ApiSign"`
}

type Response struct {
	Retcode int64       `json:"retcode"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

var (
	client  *http.Client
	timeout time.Duration = 10
	apiKey  string
	apiSerc string
	host    string = "http://api.varena.com"
)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func InitClient(key string, serc string) {
	timeout = timeout * time.Second
	apiKey = key
	apiSerc = serc
	tr := &http.Transport{
		Dial: dialTimeout,
		ResponseHeaderTimeout: time.Second * 2,
	}
	client = &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}
}

func Get(uri string, args map[string]string) (response *Response, err error) {
	resp, err := http.NewRequest("GET", fmt.Sprintf("%s:%d%s", host, 80, uri), nil)
	if err != nil {
		return nil, err
	}
	var header Header
	err = buildHeader(args, &header, uri)
	if err != nil {
		return nil, err
	}
	resp.Header.Set("Accept-ApiKey", header.AcceptApiKey)
	resp.Header.Set("Accept-ApiNonce", header.AcceptApiNonce)
	resp.Header.Set("Accept-ApiTime", header.AcceptApiTime)
	resp.Header.Set("Accept-ApiSign", header.AcceptApiSign)
	resp.Header.Set("Content-Type", "application/json")
	res, err := client.Do(resp)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println(err, string(body))
		return nil, err
	}

	return response, nil
}

func Post(uri string, args map[string]string) (response *Response, err error) {
	var keys []string
	var queryStr string
	for _, k := range keys {
		queryStr += fmt.Sprintf("%s=%s&", k, args[k])
	}
	resp, err := http.NewRequest("POST", fmt.Sprintf("%s:%d%s", host, 80, uri), strings.NewReader(queryStr))
	if err != nil {
		return nil, err
	}
	var header Header
	err = buildHeader(args, &header, uri)
	if err != nil {
		return nil, err
	}

	resp.Header.Set("Accept-ApiKey", header.AcceptApiKey)
	resp.Header.Set("Accept-ApiNonce", header.AcceptApiNonce)
	resp.Header.Set("Accept-ApiTime", header.AcceptApiTime)
	resp.Header.Set("Accept-ApiSign", header.AcceptApiSign)
	resp.Header.Set("Content-Type", "application/json")
	res, err := client.Do(resp)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func buildHeader(args map[string]string, header *Header, uri string) error {
	var queryStr string
	var argsStr string
	//sort
	if len(args) > 1 {
		var keys []string
		for k := range args {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			queryStr += fmt.Sprintf("%s=%s&", k, args[k])
		}
		queryStr = strings.Trim(queryStr, "&")
		argsStr = string([]byte(encrypt(queryStr)))
	} else {
		queryStr = ""
		argsStr = ""
	}
	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)
	nonce := string([]byte(encrypt(timestampStr)))[8:13]
	arr := []string{
		nonce,
		apiSerc,
		timestampStr,
		uri,
		argsStr,
	}
	md5Str := strings.Join(arr, "|")
	sign := encrypt(md5Str)
	header.AcceptApiKey = apiKey
	header.AcceptApiNonce = nonce
	header.AcceptApiTime = timestampStr
	header.AcceptApiSign = sign
	return nil
}

func encrypt(str string) string {
	h := md5.New()
	h.Write([]byte(string(str)))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}
