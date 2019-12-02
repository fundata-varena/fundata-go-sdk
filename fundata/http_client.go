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
	"log"
)

type Header struct {
	AcceptApiKey   string `json:"Accept-ApiKey"`
	AcceptApiNonce string `json:"Accept-ApiNonce"`
	AcceptApiTime  string `json:"Accept-ApiTime"`
	AcceptApiSign  string `json:"Accept-ApiSign"`
}

type Response struct {
	RetCode int64       `json:"retcode"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

var (
	client    *http.Client
	timeout   time.Duration = 10
	apiKey    string
	apiSecret string
	host          = "http://api.varena.com"
)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func InitClient(key string, secret string) {
	if client != nil {
		return
	}
	timeout = timeout * time.Second
	apiKey = key
	apiSecret = secret

	tr := &http.Transport{
		Dial: dialTimeout,
		ResponseHeaderTimeout: time.Second * 2,
	}

	client = &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}
}

func Get(uri string, args map[string]interface{}) (response *Response, err error) {
	fullURI := fmt.Sprintf("%s:%d%s", host, 80, uri)

	resp, err := http.NewRequest("GET", fullURI, nil)
	if err != nil {
		log.Println("New http GET request error", fullURI)
		return nil, err
	}

	if len(args) > 0 {
		q :=resp.URL.Query()
		for k, v := range args {
			q.Add(k, valueToString(v))
		}
		resp.URL.RawQuery = q.Encode()
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
		log.Println("Request GET failed", resp.URL.String(), err)
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Read GET response failed", resp.URL.String(), err)
		return nil, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println("Unmarshal GET response failed", resp.URL.String(), err, string(body))
		return nil, err
	}

	return response, nil
}

func Post(uri string, args map[string]interface{}) (response *Response, err error) {
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
		log.Println("Request POST failed", uri, err)
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Read POST response failed", uri, err)
		return nil, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println("Unmarshal POST response failed", uri, err, string(body))
		return nil, err
	}

	return response, nil
}

func valueToString(val  interface{}) string {
	var str string

	switch t := val.(type) {
	case int8:
	case int16:
	case int32:
	case int64:
	case uint8:
	case uint16:
	case uint32:
	case uint64:
	case int:
		str = fmt.Sprintf("%d", val)
	case float32:
	case float64:
		str = fmt.Sprintf("%f", val)
	case bool:
		str = fmt.Sprintf("%t", val)
	case string:
		str = fmt.Sprintf("%s", val)
	default:
		log.Println("Not supported value type", val, t)
		str = ""
	}

	return str
}

func buildHeader(args map[string]interface{}, header *Header, uri string) error {
	var queryStr string
	var argsStr string

	// sort
	if len(args) > 0 {
		var keys []string
		for k := range args {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		queryStrs := make([]string, 0)
		for _, k := range keys {
			queryStrs = append(queryStrs, fmt.Sprintf("%s=%s", k, valueToString(args[k])))
		}
		queryStr = strings.Join(queryStrs, "&")

		argsStr = queryStr
	} else {
		queryStr = ""
		argsStr = ""
	}

	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)
	nonce := string([]byte(encrypt(timestampStr)))[8:13]
	arr := []string{
		nonce,
		apiSecret,
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
