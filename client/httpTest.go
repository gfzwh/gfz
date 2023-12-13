package client

import (
	"fmt"
	"testing"
)

func TestHttpGet(t *testing.T) {
	request := NewHttpClient()
	err := request.NewRequest("http://127.0.0.1:9200", nil)
	if nil != err {
		fmt.Println("NewRequestFail, ", err)

		return
	}

	response, err := request.Get()
	if nil != err {
		fmt.Println("Http Get Fail, ", err)

		return
	}

	fmt.Println("Http Get Success: ", string(response))

	return
}

func TestHttpPut(t *testing.T) {
	body := make(map[string]map[string]interface{})
	body["settings"] = make(map[string]interface{})
	body["settings"]["number_of_shards"] = 3
	body["settings"]["number_of_replicas"] = 2

	request := NewHttpClient()
	err := request.NewRequest("http://127.0.0.1:9200/index", body)
	if nil != err {
		fmt.Println("NewRequest Fail, ", err)

		return
	}

	request.Header("Content-Type", "application/json")
	response, err := request.Put()
	if nil != err {
		fmt.Println("Http Put Fail, ", err)

		return
	}

	fmt.Println("Http Put Success: ", string(response))

	return
}

func TestHttpDelete(t *testing.T) {
	request := NewHttpClient()
	err := request.NewRequest("http://127.0.0.1:9200/index", nil)
	if nil != err {
		fmt.Println("NewRequest Fail, ", err)

		return
	}

	response, err := request.Delete()
	if nil != err {
		fmt.Println("Http Delete Fail, ", err)

		return
	}

	fmt.Println("Http Delete Success: ", string(response))

	return
}

func TestHttpPost(t *testing.T) {
	body := make(map[string]map[string]interface{})
	body["settings"] = make(map[string]interface{})
	body["settings"]["number_of_shards"] = 3
	body["settings"]["number_of_replicas"] = 2

	request := NewHttpClient()
	err := request.NewRequest("http://127.0.0.1:9200/index", body)
	if nil != err {
		fmt.Println("NewRequest Fail, ", err)

		return
	}

	request.Header("Content-Type", "application/json")
	response, err := request.Post()
	if nil != err {
		fmt.Println("Http Post Fail, ", err)

		return
	}

	fmt.Println("Http Post Success: ", string(response))

	return
}
