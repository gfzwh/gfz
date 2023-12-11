package client

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/shockerjue/gfz/common"
	"github.com/shockerjue/gfz/proto"
	"github.com/shockerjue/gfz/tcp"
	"github.com/shockerjue/gfz/zzlog"
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

func (c *Client) call(conn *net.TCPConn, rpc string, packet []byte, opts ...CallOption) ([]byte, error) {
	startAt := time.Now().UnixMilli()
	defer func() {
		zzlog.Warnf("%s call cost %dms\n", rpc, time.Now().UnixMilli()-startAt)
	}()
	opt := initOpt(opts...)

	serialNumber := serialNumber()
	if opt.onlyCall {
		serialNumber = 0
	}

	data := &proto.MessageReq{
		SerialNumber: serialNumber,
		Headers:      make(map[string]string),
		RpcId:        int64(common.GenMethodNum(rpc)),
		Packet:       packet,
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
	Pools().WaitReq[serialNumber] = cc
	Pools().wrw.Unlock()

	_, err = tcp.WriteToConnections(conn, req)
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
		packet = Pools().WaitReq[serialNumber].Packet
		Pools().WaitReq[serialNumber].Packet = nil
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

	conn, err := Pools().connectByName(req.name)
	if nil != err {
		return nil, err
	}
	c.t = conn

	res, err := c.call(c.t, req.m, packet, opts...)
	return res, nil
}
