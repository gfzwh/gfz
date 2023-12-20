package socket

import "net"

type Request struct {
	*net.TCPListener
	*net.TCPConn
	length int
	packet []byte
}

func (r *Request) Packet() []byte {
	return r.packet
}

func (r *Request) Length() int {
	return r.length
}
