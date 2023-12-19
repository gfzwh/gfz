package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/gfzwh/gfz/common"
	"github.com/gfzwh/gfz/proto"
	"github.com/gfzwh/gfz/socket"
	"github.com/gfzwh/gfz/zzlog"
	"go.uber.org/zap"
)

type Request struct {
	name string
	in   Message
	m    string
}

type Client struct {
	t *net.TCPConn
}

func (c *Client) NewRequest(name string, m string, in Message) *Request {
	return &Request{
		name: name,
		in:   in,
		m:    m,
	}
}

func (c *Client) call(res *socket.Response, rpc string, packet []byte, opts ...CallOption) ([]byte, error) {
	startAt := time.Now().UnixMilli()
	defer func() {
		zzlog.Warnw("call success", zap.Any(rpc, fmt.Sprintf("%dms", time.Now().UnixMilli()-startAt)))
	}()

	opt := initOpt(opts...)
	Sid := Sid()
	if opt.onlyCall {
		Sid = 0
	}

	data := &proto.MessageReq{
		Sid:     Sid,
		Headers: make(map[string]string),
		RpcId:   int64(common.GenMethodNum(rpc)),
		Packet:  packet,
	}

	req, err := data.Marshal()
	if nil != err {
		return nil, err
	}

	waitCh := make(chan int)
	cc := &CallCond{
		Ch: waitCh,
	}

	Pools().wrw.Lock()
	Pools().WaitReq[Sid] = cc
	Pools().wrw.Unlock()

	_, err = res.Write(req)
	if nil != err {
		return nil, err
	}

	if opt.onlyCall {
		return make([]byte, 0), nil
	}

	// 进行异步处理
	select {
	case <-time.After(time.Second * time.Duration(opt.timeout)):
		return nil, errors.New("Wait fail , timeout")

	case <-waitCh:
		Pools().wrw.RLock()
		packet = Pools().WaitReq[Sid].Packet
		Pools().WaitReq[Sid].Packet = nil
		Pools().wrw.RUnlock()

		break
	}

	return packet, nil
}

func (c *Client) Call(ctx context.Context, req *Request, in Message, opts ...CallOption) ([]byte, error) {
	packet, err := req.in.Marshal()
	if nil != err {
		return nil, err
	}

	response, err := Pools().connectByName(req.name)
	if nil != err {
		return nil, err
	}

	res, err := c.call(response, req.m, packet, opts...)
	return res, nil
}
