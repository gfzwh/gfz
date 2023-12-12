package protocol

import (
	"context"
	"errors"
	"tcpconnode/gfz/client"
	"tcpconnode/gfz/server"
)

// 这里是生成的client代码
type UserService struct {
	name string
}

func NewUserService(s2sname string) *UserService {
	return &UserService{
		name: s2sname,
	}
}

func (c *UserService) CreateUser(ctx context.Context, in *CreateUserReq) (resp *CreateUserResp, err error) {
	if nil == in {
		return nil, errors.New("CreateUser req is nil")
	}

	client := client.Client{}
	req := client.NewRequest(c.name, "UserService.CreateUser", in)
	resp = new(CreateUserResp)
	res, err := client.Call(ctx, req, in)
	if nil != err {
		return nil, err
	}

	resp = &CreateUserResp{}
	err = resp.Unmarshal(res)
	return resp, err
}

// 下面是生成的服务代码
type UserServiceHandler interface {
	CreateUser(context.Context, *CreateUserReq) (*CreateUserResp, error)
}

func RegisterUserServiceHandler(s *server.Server, hdlr UserServiceHandler) error {
	type userService interface {
		CreateUser(ctx context.Context, in []byte) (out []byte, err error)
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
