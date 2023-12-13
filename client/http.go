package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type HttpClient struct {
	Request  *http.Request
	Response *http.Response

	Timeout int
}

func NewHttpClient() *HttpClient {
	obj := &HttpClient{
		Timeout: 3,
	}

	return obj
}

// 设置请求的头部
//
// @param key
// @param value
func (this *HttpClient) Header(key, value string) {
	if nil == this.Request {
		return
	}

	this.Request.Header.Add(key, value)
	return
}

// 创建一个请求
//
// @param url 		请求的地址
// @param data 		请求的数据
//
// @return err
func (this *HttpClient) NewRequest(url string, data interface{}) (err error) {
	if nil == data {
		req, err1 := http.NewRequest("", url, nil)
		if nil != err {
			err = err1

			return
		}

		this.Request = req

		return
	}

	t := fmt.Sprintf("%T", data)
	switch t {
	case "string":
		body := bytes.NewReader([]byte(data.(string)))
		req, err := http.NewRequest("", url, body)
		if nil != err {
			break
		}

		this.Request = req

		break

	case "[]uint8":
		body := bytes.NewReader(data.([]byte))
		req, err := http.NewRequest("", url, body)
		if nil != err {
			break
		}

		this.Request = req

		break

	default:
		if strings.HasPrefix(t, "map[") {
			b, err := json.Marshal(data)
			if nil != err {
				break
			}

			body := bytes.NewReader(b)
			req, err := http.NewRequest("", url, body)
			if nil != err {
				break
			}

			this.Request = req
		}

		break
	}

	return
}

// 发起Http的get请求
//
// @return data,err
func (this *HttpClient) Get() (r []byte, err error) {
	if nil == this.Request {
		err = errors.New("Http Request didn't create, please call NewRequest to create!")

		return
	}

	this.Request.Method = "GET"
	client := &http.Client{
		Timeout: time.Duration(this.Timeout) * time.Second,
	}

	res, err := client.Do(this.Request)
	if nil != err {
		return
	}
	defer res.Body.Close()
	if 200 > res.StatusCode || 299 < res.StatusCode {
		err = fmt.Errorf("http request fail, statusCode: %d", res.StatusCode)

		return
	}

	r, err = io.ReadAll(res.Body)
	if nil != err {
		return
	}

	this.Response = res
	return
}

// 发起Http的PUT请求
//
// @return data,err
func (this *HttpClient) Put() (r []byte, err error) {
	if nil == this.Request {
		err = errors.New("Http Request didn't create, please call NewRequest to create!")

		return
	}

	this.Request.Method = "PUT"
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	res, err := client.Do(this.Request)
	if nil != err {
		return
	}

	defer res.Body.Close()
	r, err = ioutil.ReadAll(res.Body)

	return
}

// 发起Http的delete请求
//
// @return data,err
func (this *HttpClient) Delete() (r []byte, err error) {
	if nil == this.Request {
		err = errors.New("Http Request didn't create, please call NewRequest to create!")

		return
	}

	this.Request.Method = "DELETE"
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	res, err := client.Do(this.Request)
	if nil != err {
		return
	}

	defer res.Body.Close()
	r, err = ioutil.ReadAll(res.Body)

	return
}

// 发起Http的post请求
//
// @return data,err
func (this *HttpClient) Post() (r []byte, err error) {
	if nil == this.Request {
		err = errors.New("Http Request didn't create, please call NewRequest to create!")

		return
	}

	this.Request.Method = "POST"
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	res, err := client.Do(this.Request)
	if nil != err {
		return
	}

	defer res.Body.Close()
	r, err = ioutil.ReadAll(res.Body)

	return
}

// @param url 请求的地址
// @param data 	请求体
//
// @return data,err
func PostBase64(url string, data interface{}) (r []byte, err error) {
	b, err := json.Marshal(data)
	if err != nil {
		return
	}
	b1 := base64.StdEncoding.EncodeToString(b)
	body := bytes.NewReader([]byte(b1))
	res, err := http.Post(url, "application/json", body)
	if err != nil {
		return
	}
	defer res.Body.Close()

	r, err = io.ReadAll(res.Body)
	return
}
