package controller

import (
	"context"
	"tcpconnode/protocol"

	"github.com/shockerjue/gfz/zzlog"
)

func (this *controller) CreateUser(ctx context.Context, req *protocol.CreateUserReq) (resp *protocol.CreateUserResp, err error) {
	resp = &protocol.CreateUserResp{
		Code:  200,
		Msg:   "CreateUser Success " + req.Username,
		Extra: make(map[string]string),
	}

	return
}

func (this *controller) UserInfo(ctx context.Context, req *protocol.UserInfoReq) (resp *protocol.UserInfoResp, err error) {
	resp = &protocol.UserInfoResp{
		Code:  200,
		Msg:   "UserInfo Success " + req.Username,
		Extra: make(map[string]string),
	}

	zzlog.Infof("userInfo res:%v\n", resp)
	return
}
