package tcp

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/gfzwh/gfz/zzlog"
)

var (
	DefaultMaxMessageSize = int(1 << 20)
)

func byteArrayToUInt32(bytes []byte) (result int64, bytesRead int) {
	return binary.Varint(bytes)
}

func intToByteArray(value int64, bufferSize int) []byte {
	toWriteLen := make([]byte, bufferSize)
	binary.PutVarint(toWriteLen, value)
	return toWriteLen
}

type ListenCb func(context.Context, *net.TCPListener) error
type RecvCb func(context.Context, *net.TCPConn, int, []byte) error
type ConnectCb func(context.Context, *net.TCPConn) error
type CloseCb func(context.Context, *net.TCPConn) error
type CmdCb func(context.Context, *net.TCPConn, string) error

type TCPListener struct {
	socket          *net.TCPListener
	address         string
	headerByteSize  int
	maxMessageSize  int
	listenCb        ListenCb
	connectCb       ConnectCb
	recvCb          RecvCb
	closeCb         CloseCb
	shutdownChannel chan struct{}
	shutdownGroup   *sync.WaitGroup
}

type TCPListenerConfig struct {
	MaxMessageSize int
	EnableLogging  bool
	Address        string
	ListenCb       ListenCb
	ConnectCb      ConnectCb
	RecvCb         RecvCb
	CloseCb        CloseCb
}

func ListenTCP(cfg TCPListenerConfig) (*TCPListener, error) {
	maxMessageSize := DefaultMaxMessageSize
	// 0 is the default, and the message must be atleast 1 byte large
	if cfg.MaxMessageSize != 0 {
		maxMessageSize = cfg.MaxMessageSize
	}
	btl := &TCPListener{
		maxMessageSize:  maxMessageSize,
		headerByteSize:  4, // 4byte(int32)
		listenCb:        cfg.ListenCb,
		connectCb:       cfg.ConnectCb,
		recvCb:          cfg.RecvCb,
		closeCb:         cfg.CloseCb,
		shutdownChannel: make(chan struct{}),
		address:         cfg.Address,
		shutdownGroup:   &sync.WaitGroup{},
	}

	if err := btl.openSocket(); err != nil {
		return nil, err
	}

	return btl, nil
}

func (btl *TCPListener) blockListen() error {
	for {
		conn, err := btl.socket.AcceptTCP()
		if err != nil {
			select {
			case <-btl.shutdownChannel:
				return nil
			default:
			}
		} else {

			if nil != btl.connectCb {
				btl.connectCb(context.TODO(), conn)
			}

			go handleListenedConn(conn, btl.headerByteSize, btl.maxMessageSize, btl.recvCb, btl.closeCb)
		}
	}
}

func (btl *TCPListener) openSocket() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", btl.address)
	if err != nil {
		return err
	}
	receiveSocket, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	if nil != btl.listenCb {
		btl.listenCb(context.TODO(), receiveSocket)
	}

	btl.socket = receiveSocket
	return err
}

func (btl *TCPListener) StartListening() error {
	return btl.blockListen()
}

func (btl *TCPListener) Close() {
	close(btl.shutdownChannel)
	btl.shutdownGroup.Wait()
}

func (btl *TCPListener) StartListeningAsync() error {
	var err error
	go func() {
		err = btl.blockListen()
	}()
	return err
}

func handleListenedConn(conn *net.TCPConn, headerByteSize int, maxMessageSize int, rcb RecvCb, ccb CloseCb) {
	headerBuffer := make([]byte, headerByteSize)
	dataBuffer := make([]byte, maxMessageSize)
	defer func() {
		if err := recover(); nil != err {
			zzlog.Errorf("handleListenedConn except, err:%v\n", err)
		}

		if nil != conn {
			zzlog.Errorf("Address %s: Client closed connection\n", conn.RemoteAddr())
			conn.Close()
		}

		if nil != ccb {
			ccb(context.TODO(), conn)
		}

		return
	}()

	for {
		var headerReadError error
		var totalHeaderBytesRead = 0
		var bytesRead = 0
		// First, read the number of bytes required to determine the message length
		for totalHeaderBytesRead < headerByteSize && headerReadError == nil {
			// While we haven't read enough yet, pass in the slice that represents where we are in the buffer
			bytesRead, headerReadError = readFromConnection(conn, headerBuffer[totalHeaderBytesRead:])
			totalHeaderBytesRead += bytesRead
		}
		if headerReadError != nil {
			if headerReadError != io.EOF {
				// Log the error we got from the call to read
				zzlog.Errorf("Error when trying to read from address %s. Tried to read %d, actually read %d. Underlying error: %s\n", conn.RemoteAddr(), headerByteSize, totalHeaderBytesRead, headerReadError)
			} else {
				// Client closed the conn
				zzlog.Errorf("Address %s: Client closed connection during header read. Underlying error: %s\n", conn.RemoteAddr(), headerReadError)
			}

			return
		}
		// Now turn that buffer of bytes into an integer - represnts size of message body
		msgLength, bytesParsed := byteArrayToUInt32(headerBuffer)
		iMsgLength := int(msgLength)
		// Not sure what the correct way to handle these errors are. For now, bomb out
		if bytesParsed == 0 {
			// "Buffer too small"
			zzlog.Errorf("Address %s: 0 Bytes parsed from header. Underlying error: %s\n", conn.RemoteAddr(), headerReadError)
			conn.Close()
			return
		} else if bytesParsed < 0 {
			// "Buffer overflow"
			zzlog.Errorf("Address %s: Buffer Less than zero bytes parsed from header. Underlying error: %s\n", conn.RemoteAddr(), headerReadError)
			return
		}
		var dataReadError error
		var totalDataBytesRead = 0
		bytesRead = 0

		// 读取消息，直到满足消息长度
		for totalDataBytesRead < iMsgLength && dataReadError == nil {
			bytesRead, dataReadError = readFromConnection(conn, dataBuffer[totalDataBytesRead:iMsgLength])
			totalDataBytesRead += bytesRead
		}

		if dataReadError != nil {
			if dataReadError != io.EOF {
				// log the error from the call to read
				zzlog.Errorf("Address %s: Failure to read from connection. Was told to read %d by the header, actually read %d. Underlying error: %s\n", conn.RemoteAddr(), msgLength, totalDataBytesRead, dataReadError)
			} else {
				// The client wrote the header but closed the connection
				zzlog.Errorf("Address %s: Client closed connection during data read. Underlying error: %s\n", conn.RemoteAddr(), dataReadError)
			}

			return
		}

		// 如果读取消息没有错误，就调用回调函数
		if totalDataBytesRead > 0 && (dataReadError == nil || (dataReadError != nil && dataReadError == io.EOF)) {
			// 防止粘包
			packet := make([]byte, iMsgLength)
			copy(packet, dataBuffer[:iMsgLength])

			go func(packet []byte) {
				err := rcb(context.TODO(), conn, iMsgLength, packet)
				if err != nil {
					zzlog.Errorf("Error in Callback: %s\n", err)
				}
			}(packet)
		}
	}
}

// Handles reading from a given connection.
func readFromConnection(reader *net.TCPConn, buffer []byte) (int, error) {
	// This fills the buffer
	bytesLen, err := reader.Read(buffer)
	// Output the content of the bytes to the queue
	if bytesLen == 0 {
		if err != nil && err == io.EOF {
			// "End of individual transmission"
			// We're just done reading from that conn
			return bytesLen, err
		}
	}

	if err != nil {
		//"Underlying network failure?"
		// Not sure what this error would be, but it could exist and i've seen it handled
		// as a general case in other networking code. Following in the footsteps of (greatness|madness)
		return bytesLen, err
	}
	// Read some bytes, return the length
	return bytesLen, nil
}

// 读取数据
func ReadFromTcp(conn *net.TCPConn, rcb RecvCb, ccb CloseCb) (err error) {
	handleListenedConn(conn, 4, DefaultMaxMessageSize, rcb, ccb)

	return
}

func WriteToConnections(conn *net.TCPConn, packet []byte) (n int, err error) {
	if 0 == len(packet) {
		return
	}

	if nil == conn {
		fmt.Println("WriteToConnections conn is nil")

		return
	}

	msgLenHeader := intToByteArray(int64(len(packet)), 4)
	toWrite := append(msgLenHeader, packet...)

	toWriteLen := len(toWrite)
	var writeError error
	var totalBytesWritten = 0
	var bytesWritten = 0
	for totalBytesWritten < toWriteLen && writeError == nil {
		bytesWritten, writeError = conn.Write(toWrite[totalBytesWritten:])
		totalBytesWritten += bytesWritten
	}

	err = writeError
	n = totalBytesWritten
	return
}
