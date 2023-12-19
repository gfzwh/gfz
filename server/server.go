package server

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bilibili/discovery/naming"
	"github.com/gfzwh/gfz/client"
	"github.com/gfzwh/gfz/common"
	"github.com/gfzwh/gfz/proto"
	"github.com/gfzwh/gfz/registry"
	"github.com/gfzwh/gfz/socket"
	"github.com/gfzwh/gfz/zzlog"
	"go.uber.org/zap"

	"github.com/StabbyCutyou/buffstreams"
)

type RidItem struct {
	call reflect.Value
	name string
}

type Server struct {
	rw         sync.RWMutex
	rpcMap     map[uint64]RidItem
	cancelFunc context.CancelFunc
	registry   *registry.Registry
	n          *node

	reqs  int64 // 当前正在处理的请求
	conns int64 // 当前连接的数量
}

func NewServer(registry *registry.Registry, n *node) *Server {
	return &Server{
		rw:       sync.RWMutex{},
		rpcMap:   make(map[uint64]RidItem),
		registry: registry,
		n:        n,
	}
}

func Limit(ctx context.Context) {

}

func (this *Server) Use() {

}

func (this *Server) incReq() int64 {
	return atomic.AddInt64(&this.reqs, 1)
}

func (this *Server) decReq() int64 {
	return atomic.AddInt64(&this.reqs, -1)
}

func (this *Server) incConn() int64 {
	return atomic.AddInt64(&this.conns, 1)
}

func (this *Server) decConn() int64 {
	return atomic.AddInt64(&this.conns, -1)
}

func (s *Server) listen(ctx context.Context, req *socket.Request) error {
	s.register(req.Addr().String())
	client.Pools().Registry(registry.NewRegistry(
		registry.Url(s.registry.Url()),
		registry.Zone(s.registry.Zone()),
		registry.Env(s.registry.Env()),
		registry.Host(s.registry.Host())))

	zzlog.Infow("Server.listen called", zap.String("addr", req.Addr().String()))
	return nil
}

func (this *Server) connect(ctx context.Context, req *socket.Request) error {
	this.incConn()
	zzlog.Infow("Server.connect called", zap.String("from", req.RemoteAddr().String()))

	return nil
}

func (this *Server) closed(ctx context.Context, req *socket.Request) error {
	this.decConn()
	zzlog.Infow("Server.closed called", zap.String("from", req.RemoteAddr().String()))

	return nil
}

// method_num|data
func (this *Server) recv(ctx context.Context, request *socket.Request, response *socket.Response) error {
	statAt := time.Now().UnixMilli()
	msg := &proto.MessageReq{}
	err := msg.Unmarshal(request.Data())
	if nil != err {
		return err
	}

	var item RidItem
	this.rw.RLock()
	item = this.rpcMap[uint64(msg.GetRpcId())]
	this.rw.RUnlock()

	reqCount := this.incReq()
	ctx = context.WithValue(ctx, "reqCount", reqCount)

	defer func() {
		reqCount = this.decReq()
		zzlog.Debugw("Recv from client",
			zap.Int64("Sid", msg.Sid),
			zap.Int64("reqCount", reqCount),
			zap.Int64("conns", this.conns),
			zap.String("cost", fmt.Sprintf("%dms", time.Now().UnixMilli()-statAt)))
	}()

	res := &proto.MessageResp{
		Sid:     msg.Sid,
		Headers: msg.Headers,
		Code:    0,
	}

	if common.ValueEmpty(item.call) {
		return errors.New("Not support called!")
	}

	// 处理流量限制、熔段

	params := make([]reflect.Value, 2)
	params[0] = reflect.ValueOf(context.TODO())
	params[1] = reflect.ValueOf(msg.Packet)
	ret := item.call.Call(params)

	// 不需要响应的直接返回
	if 0 == msg.Sid {
		return nil
	}

	res.Packet = ret[0].Interface().([]byte)
	outErr := ret[1].Interface()
	if nil != outErr {
		err = outErr.(error)
		return err
	}

	packet, err := res.Marshal()
	if nil != err {
		return err
	}

	response.Write(packet)
	return nil
}

func (s *Server) NewHandler(instance interface{}) error {
	structName := reflect.TypeOf(instance).Name()
	t := reflect.TypeOf(instance)

	value := reflect.ValueOf(instance)

	// 遍历结构体的方法
	for i := 0; i < t.NumMethod(); i++ {
		// 生成请求rid
		method := t.Method(i)
		rid := common.GenMethodNum(fmt.Sprintf("%s.%s", structName, method.Name))

		methodValue := value.Method(i)

		s.rw.Lock()
		s.rpcMap[rid] = RidItem{
			call: methodValue,
			name: fmt.Sprintf("%s.%s", structName, method.Name),
		}

		s.rw.Unlock()
	}

	return nil
}

func (s *Server) register(addr string) {
	// 下面是discovery节点的信息
	conf := &naming.Config{
		Nodes:  s.registry.Nodes(), // NOTE: 配置种子节点(1个或多个)，client内部可根据/discovery/nodes节点获取全部node(方便后面增减节点)
		Region: s.registry.Region(),
		Zone:   s.registry.Zone(),
		Host:   s.registry.Host(),
		Env:    s.registry.Env(),
	}

	dis := naming.New(conf)

	// 服务的节点信息
	ins := &naming.Instance{
		Zone:     s.n.opts.zone,
		Env:      s.n.opts.env,
		AppID:    s.n.opts.name, // 服务名，如 usernode
		Addrs:    []string{fmt.Sprintf("tcp://%s", addr)},
		LastTs:   time.Now().Unix(),
		Metadata: map[string]string{"weight": "10"},
	}

	s.cancelFunc, _ = dis.Register(ins)
}

func (s *Server) Release() {
	s.cancelFunc()
}

func (s *Server) Run(opts ...HandlerOption) {
	config := &options{
		bind: "0.0.0.0",
		port: 0,
	}

	for _, o := range opts {
		o(config)
	}

	btl, err := socket.ListenTCP(
		socket.MaxMessageSize(1<<20),
		socket.EnableLogging(true),
		socket.Address(buffstreams.FormatAddress(config.bind, strconv.Itoa(config.port))),
		socket.Listen(s.listen),
		socket.Connect(s.connect),
		socket.Closed(s.closed),
		socket.Recv(s.recv),
	)
	if err != nil {
		zzlog.Errorw("ListenTCP error", zap.Error(err))

		return
	}
	defer btl.Close()

	err = btl.StartListeningAsync()
	if nil != err {
		zzlog.Errorw("StartListening error", zap.Error(err))

		return
	}
}
