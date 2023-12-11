package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shockerjue/gfz/common"
	"github.com/shockerjue/gfz/proto"
	"github.com/shockerjue/gfz/tcp"
	"github.com/shockerjue/gfz/zzlog"

	"github.com/StabbyCutyou/buffstreams"
)

type RidItem struct {
	call reflect.Value
	name string
}

type Server struct {
	rw     sync.RWMutex
	rpcMap map[uint64]RidItem

	reqs  int64 // 当前正在处理的请求
	conns int64 // 当前连接的数量
}

func NewServer() *Server {
	return &Server{
		rw:     sync.RWMutex{},
		rpcMap: make(map[uint64]RidItem),
	}
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

func (this *Server) listenCb(ctx context.Context, tcp *net.TCPListener) error {
	zzlog.Infof("LCallback ----------  addr:%s\n", tcp.Addr().String())

	return nil
}

func (this *Server) connectCb(ctx context.Context, tcp *net.TCPConn) error {
	this.incConn()
	zzlog.Infof("ConnectCb ---------- from: %s\n", tcp.RemoteAddr())

	return nil
}

func (this *Server) closeCb(ctx context.Context, tcp *net.TCPConn) error {
	this.decConn()
	zzlog.Infof("CloseCb ---------- from: %s\n", tcp.RemoteAddr())

	return nil
}

// method_num|data
func (this *Server) recvCb(ctx context.Context, client *net.TCPConn, iMsgLength int, data []byte) error {
	statAt := time.Now().UnixMilli()

	msg := &proto.MessageReq{}
	err := msg.Unmarshal(data)
	if nil != err {
		return err
	}

	var item RidItem
	this.rw.RLock()
	item = this.rpcMap[uint64(msg.GetRpcId())]
	this.rw.RUnlock()

	reqCount := this.incReq()
	if 0 < reqCount {

	}

	zzlog.Infof("Recv from client, SerialNumber:%d	req_number:%d	conns:%d\n", msg.SerialNumber, reqCount, this.conns)
	defer func() {
		this.decReq()
		zzlog.Warnf("%s called ,cost %dms\n", item.name, time.Now().UnixMilli()-statAt)
	}()

	res := &proto.MessageResp{
		SerialNumber: msg.SerialNumber,
		Headers:      make(map[string]string),
	}

	if common.ValueEmpty(item.call) {
		return errors.New("Not support called!")
	}

	params := make([]reflect.Value, 2)
	params[0] = reflect.ValueOf(context.TODO())
	params[1] = reflect.ValueOf(msg.Packet)
	ret := item.call.Call(params)

	// 不需要响应的直接返回
	if 0 == msg.SerialNumber {
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

	tcp.WriteToConnections(client, packet)
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

func (s *Server) Run(opts ...HandlerOption) {
	config := &options{
		bind: "0.0.0.0",
		port: 0,
	}

	for _, o := range opts {
		o(config)
	}

	cfg := tcp.TCPListenerConfig{
		MaxMessageSize: 1 << 20,
		EnableLogging:  true,
		Address:        buffstreams.FormatAddress(config.bind, strconv.Itoa(config.port)),
		ListenCb:       s.listenCb,
		ConnectCb:      s.connectCb,
		CloseCb:        s.closeCb,
		RecvCb:         s.recvCb,
	}

	btl, err := tcp.ListenTCP(cfg)
	if err != nil {
		zzlog.Errorf("ListenTCP error {%v}\n", err)

		return
	}
	defer btl.Close()

	err = btl.StartListeningAsync()
	if nil != err {
		zzlog.Errorf("StartListening error {%v}\n", err)

		return
	}
}
