package client

import (
	"net"
	"sync"
	"sync/atomic"
)

const (
	CONNECT_POOLS_SIZE = 5
)

var counter int64

func Sid() int64 {
	return atomic.AddInt64(&counter, 1)
}

var instance *pools
var once sync.Once

type client struct {
	T       *net.TCPConn
	Name    string
	Svrname string
}

type CallCond struct {
	Ch     chan int
	Packet []byte
}

type Message interface {
	Unmarshal(dAtA []byte) error
	Marshal() (dAtA []byte, err error)
}
