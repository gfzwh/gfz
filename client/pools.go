package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/gfzwh/gfz/common"
	"github.com/gfzwh/gfz/proto"
	"github.com/gfzwh/gfz/registry"
	"github.com/gfzwh/gfz/tcp"
	"github.com/gfzwh/gfz/zzlog"
)

type pools struct {
	rw      sync.RWMutex
	wrw     sync.RWMutex
	WaitReq map[int64]*CallCond
	clients map[string][]*client

	r    *registry.Registry
	opts *Options
}

func Pools() *pools {
	once.Do(func() {
		instance = &pools{
			clients: make(map[string][]*client),
			WaitReq: make(map[int64]*CallCond),
			r:       registry.NewRegistry(),
		}
	})

	return instance
}

func (p *pools) Init(opts ...CallOption) {
	p.opts = initOpt(opts...)
}

func (p *pools) connect(svrname, name string) (t *net.TCPConn, err error) {
	addr, err := p.r.GetNodeInfo(svrname, p.opts.zone, p.opts.env, p.opts.host)
	if nil != err {
		return
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return
	}
	t, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return
	}

	go tcp.ReadFromTcp(t, func(ctx context.Context, t *net.TCPConn, i int, b []byte) error {
		// 接收响应，并分发
		msg := &proto.MessageResp{}
		err := msg.Unmarshal(b)
		if nil != err {
			return nil
		}

		// 异步接收到响应
		serialNumber := msg.SerialNumber
		Pools().wrw.RLock()
		if _, ok := Pools().WaitReq[serialNumber]; ok {
			cc := Pools().WaitReq[serialNumber]
			cc.Packet = msg.Packet
			cc.Ch <- 0
		}
		Pools().wrw.RUnlock()
		zzlog.Infof("Recv from server, SerialNumber: %d\n", serialNumber)

		return nil
	}, func(ctx context.Context, t *net.TCPConn) error {
		index := -1

		clients := make([]*client, 0)
		p.rw.RLock()
		for k, v := range p.clients[svrname] {
			if v.Name == name {
				index = k

				break
			}
			clients = append(clients, &client{
				T:       v.T,
				Svrname: v.Svrname,
				Name:    v.Name,
			})
		}
		p.rw.RUnlock()

		if -1 != index {
			p.rw.Lock()
			p.clients[svrname] = clients
			p.rw.Unlock()
		}

		return nil
	})
	return
}

func (p *pools) connectByName(svrname string) (t *net.TCPConn, err error) {
	p.rw.Lock()
	defer p.rw.Unlock()

	genCon := func(num int) []*client {
		clients := make([]*client, 0)
		for i := 0; i < num; i++ {
			name := common.GenUid()
			conn, err := p.connect(svrname, name)
			if nil != err {
				zzlog.Errorf("pools.connect error, err:%v\n", err)

				return nil
			}

			clients = append(clients, &client{
				T:       conn,
				Svrname: svrname,
				Name:    name,
			})
		}

		return clients
	}

	if 0 == len(p.clients[svrname]) {
		p.clients[svrname] = genCon(CONNECT_POOLS_SIZE)
	}

	if 0 == len(p.clients[svrname]) {
		err = errors.New(fmt.Sprintf("%s didn't more node!", svrname))

		return
	}
	for i := len(p.clients[svrname]); i < CONNECT_POOLS_SIZE; i++ {
		clients := genCon(CONNECT_POOLS_SIZE - len(p.clients[svrname]))

		p.clients[svrname] = append(p.clients[svrname], clients...)
	}

	t = p.clients[svrname][0].T
	return
}
