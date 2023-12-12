// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: protocol/usernode.proto

/*
Package protocol is a generated protocol buffer package.

It is generated from these files:

	protocol/usernode.proto

It has these top-level messages:

	Authorize
	CreateUserReq
	CreateUserResp
	UserInfoReq
	UserInfoResp
*/
package protocol

import (
	"errors"
	client "github.com/shockerjue/gfz/client"
	server "github.com/shockerjue/gfz/server"
	context "context"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.

// Client API for UserService service

type UserService interface {
	CreateUser(ctx context.Context, in *CreateUserReq, opts ...client.CallOption) (out *CreateUserResp, err error)
	UserInfo(ctx context.Context, in *UserInfoReq, opts ...client.CallOption) (out *UserInfoResp, err error)
}

type userService struct {
	serviceName string
}

func NewUserService(serviceName string) UserService {
	if len(serviceName) == 0 {
		serviceName = "protocol"
	}
	return &userService{
		serviceName: serviceName,
	}
}

func (c *userService) CreateUser(ctx context.Context, in *CreateUserReq, opts ...client.CallOption) (out *CreateUserResp, err error) {
	if nil == in {
		err = errors.New("UserService.CreateUser req is nil")
		return
	}
	client := client.Client{}
	req := client.NewRequest(c.serviceName, "UserService.CreateUser", in)
	out = new(CreateUserResp)
	res, err := client.Call(ctx, req, in, opts...)
	if err != nil {
		return
	}
	err = out.Unmarshal(res)
	return
}

func (c *userService) UserInfo(ctx context.Context, in *UserInfoReq, opts ...client.CallOption) (out *UserInfoResp, err error) {
	if nil == in {
		err = errors.New("UserService.UserInfo req is nil")
		return
	}
	client := client.Client{}
	req := client.NewRequest(c.serviceName, "UserService.UserInfo", in)
	out = new(UserInfoResp)
	res, err := client.Call(ctx, req, in, opts...)
	if err != nil {
		return
	}
	err = out.Unmarshal(res)
	return
}

// Server API for UserService service

type UserServiceHandler interface {
	CreateUser(context.Context, *CreateUserReq) (*CreateUserResp, error)
	UserInfo(context.Context, *UserInfoReq) (*UserInfoResp, error)
}

func RegisterUserServiceHandler(s *server.Server, hdlr UserServiceHandler, opts ...server.HandlerOption) error {
	type userService interface {
		CreateUser(ctx context.Context, in []byte) (out []byte, err error)
		UserInfo(ctx context.Context, in []byte) (out []byte, err error)
	}
	type UserService struct {
		userService
	}
	h := &userServiceHandler{hdlr}
	return s.NewHandler(UserService{h})
}

type userServiceHandler struct {
	UserServiceHandler
}

func (h *userServiceHandler) CreateUser(ctx context.Context, in []byte) (out []byte, err error) {
	var req CreateUserReq
	err = req.Unmarshal(in)
	if nil != err {
		return
	}

	var res *CreateUserResp
	res, err = h.UserServiceHandler.CreateUser(ctx, &req)
	if nil != err {
		return
	}

	out, err = res.Marshal()
	if nil != err {
		return
	}
	return
}

func (h *userServiceHandler) UserInfo(ctx context.Context, in []byte) (out []byte, err error) {
	var req UserInfoReq
	err = req.Unmarshal(in)
	if nil != err {
		return
	}

	var res *UserInfoResp
	res, err = h.UserServiceHandler.UserInfo(ctx, &req)
	if nil != err {
		return
	}

	out, err = res.Marshal()
	if nil != err {
		return
	}
	return
}