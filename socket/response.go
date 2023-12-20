package socket

import (
	"net"

	"github.com/gfzwh/gfz/zzlog"
)

type Response struct {
	*net.TCPConn
}

func (r *Response) Write(packet []byte) (n int, err error) {
	if 0 == len(packet) {
		return
	}

	if nil == r.TCPConn {
		zzlog.Warnln("WriteToConnections conn is nil")

		return
	}

	msgLenHeader := intToByteArray(int64(len(packet)), 4)
	toWrite := append(msgLenHeader, packet...)

	toWriteLen := len(toWrite)
	var writeError error
	var totalBytesWritten = 0
	var bytesWritten = 0
	for totalBytesWritten < toWriteLen && writeError == nil {
		bytesWritten, writeError = r.TCPConn.Write(toWrite[totalBytesWritten:])
		totalBytesWritten += bytesWritten
	}

	err = writeError
	n = totalBytesWritten
	return
}
