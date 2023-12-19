package socket

import (
	"context"
	"encoding/binary"
	"net"
)

type Request struct {
	*net.TCPConn
	length int
	packet []byte
}

func (r *Request) Data() []byte {
	return r.packet
}

func (r *Request) Length() int {
	return r.length
}

type Response struct {
	*net.TCPConn
}

func byteArrayToUInt32(bytes []byte) (result int64, bytesRead int) {
	return binary.Varint(bytes)
}

func intToByteArray(value int64, bufferSize int) []byte {
	toWriteLen := make([]byte, bufferSize)
	binary.PutVarint(toWriteLen, value)
	return toWriteLen
}

// socket调用的方法
type listen func(context.Context, *net.TCPListener) error
type connect func(context.Context, *net.TCPConn) error
type closed func(context.Context, *net.TCPConn) error
type recv func(context.Context, *Request, *Response) error
type CmdFunc func(context.Context, *net.TCPConn, string) error
