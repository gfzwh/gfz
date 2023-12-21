package client

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/gfzwh/gfz/common"
	"github.com/gfzwh/gfz/proto"
	"github.com/gfzwh/gfz/registry"
	"github.com/gfzwh/gfz/socket"
	"github.com/gfzwh/gfz/zzlog"
	"go.uber.org/zap"
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

func (p *pools) Registry(r *registry.Registry) {
	p.r = r
}

func (p *pools) connect(svrname, name string) (t *net.TCPConn, err error) {
	addr, err := p.r.GetNodeInfo(svrname, p.r.Zone(), p.r.Env(), p.r.Host())
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

	go socket.ReadFromTcp(t, func(ctx context.Context, req *socket.Request, r *socket.Response) error {
		// 接收响应，并分发
		msg := &proto.MessageResp{}
		err := msg.Unmarshal(req.Packet())
		if nil != err {
			return nil
		}

		if 0 != msg.Code {
			zzlog.Debugw("Recv from server fail", zap.Int32("code", msg.Code), zap.Any("headers", msg.Headers))

			return nil
		}

		// 异步接收到响应
		Sid := msg.Sid
		Pools().wrw.RLock()
		if _, ok := Pools().WaitReq[Sid]; ok {
			cc := Pools().WaitReq[Sid]
			cc.Packet = msg.Packet
			cc.Ch <- 0
		}
		Pools().wrw.RUnlock()
		zzlog.Debugw("Recv from server", zap.Int64("Sid", Sid))

		return nil
	}, func(ctx context.Context, req *socket.Request) error {
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

func (p *pools) connectByName(svrname string) (t *socket.Response, err error) {
	p.rw.Lock()
	defer p.rw.Unlock()

	genCon := func(num int) []*client {
		clients := make([]*client, 0)
		for i := 0; i < num; i++ {
			name := common.GenUid()
			conn, err := p.connect(svrname, name)
			if nil != err {
				zzlog.Errorw("pools.connect error", zap.Error(err))

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
	if len(p.clients[svrname]) < CONNECT_POOLS_SIZE {
		clients := genCon(CONNECT_POOLS_SIZE - len(p.clients[svrname]))

		p.clients[svrname] = append(p.clients[svrname], clients...)
	}

	seed := time.Now().UnixNano()
	src := rand.NewSource(seed)
	rng := rand.New(src)

	length := rng.Intn(len(p.clients[svrname]))
	t = &socket.Response{
		TCPConn: p.clients[svrname][length].T,
	}
	zzlog.Debugw("Select client connect", zap.Int("index", length))
	return
}
