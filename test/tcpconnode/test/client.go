package main

import (
	"context"
	"fmt"
	"time"

	"tcpconnode/protocol"

	"github.com/shockerjue/gfz/client"
	"github.com/shockerjue/gfz/zzlog"
)

func createUser(i int) {
	// 调用等待响应
	userService := protocol.NewUserService("UserService")
	resp, err := userService.CreateUser(context.TODO(), &protocol.CreateUserReq{
		Auth:      &protocol.Authorize{Appid: "", Appkey: ""},
		Username:  fmt.Sprintf("%s-%d", "shockerjue", i),
		Telephone: "1234567890",
		Email:     "12345@123.com",
	}, client.Timeout(3))
	if nil != err {
		zzlog.Errorf("CreateUser return error, err:%v\n", err)
	} else {
		zzlog.Infof("Create return %v\n", resp)
	}
}

func userInfo() {
	// 调用接口，切设置不进行响应
	userService := protocol.NewUserService("UserService")
	_, err := userService.UserInfo(context.TODO(), &protocol.UserInfoReq{
		Auth:     &protocol.Authorize{Appid: "", Appkey: ""},
		Username: fmt.Sprintf("%s", "shockerjue"),
	}, client.OnlyCall(true))
	if nil != err {
		zzlog.Errorf("UserInfo return error, err:%v\n", err)
	}
}

func main() {
	for i := 0; i < 100; i++ {
		go createUser(i)
		go userInfo()
	}

	time.Sleep(time.Second * 5)
	return
}
