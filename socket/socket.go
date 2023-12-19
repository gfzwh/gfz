package socket

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/gfzwh/gfz/zzlog"
	"go.uber.org/zap"
)

const (
	DefaultMaxMessageSize = int(1 << 20)
)

type TCPListener struct {
	socket          *net.TCPListener
	shutdownChannel chan struct{}
	shutdownGroup   *sync.WaitGroup

	opts *options
}

func ListenTCP(opts ...SocketOption) (*TCPListener, error) {
	cfg := initOpts(opts...)
	if cfg.maxMessageSize == 0 {
		cfg.maxMessageSize = int32(DefaultMaxMessageSize)
	}
	if cfg.headerByteSize == 0 {
		cfg.headerByteSize = 4
	}

	btl := &TCPListener{
		shutdownChannel: make(chan struct{}),
		shutdownGroup:   &sync.WaitGroup{},
		opts:            cfg,
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

			if nil != btl.opts.connect {
				btl.opts.connect(context.TODO(), conn)
			}

			go handleListenedConn(conn, int(btl.opts.headerByteSize), int(btl.opts.maxMessageSize), btl.opts.recv, btl.opts.closed)
		}
	}
}

func (btl *TCPListener) openSocket() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", btl.opts.address)
	if err != nil {
		return err
	}
	receiveSocket, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	if nil != btl.opts.listen {
		btl.opts.listen(context.TODO(), receiveSocket)
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

func handleListenedConn(conn *net.TCPConn, headerByteSize int, maxMessageSize int, rcb recv, ccb closed) {
	headerBuffer := make([]byte, headerByteSize)
	dataBuffer := make([]byte, maxMessageSize)
	defer func() {
		if err := recover(); nil != err {
			zzlog.Errorw("handleListenedConn except", zap.Error(err.(error)))
		}

		if nil != conn {
			zzlog.Errorw("Client closed connection", zap.String("Address", conn.RemoteAddr().String()))

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
				zzlog.Errorw("Error when trying to read",
					zap.String("address", conn.RemoteAddr().String()),
					zap.Int("headerByteSize", headerByteSize),
					zap.Int("totalHeaderBytesRead", totalHeaderBytesRead),
					zap.Error(headerReadError))
			} else {
				// Client closed the conn
				zzlog.Errorw("Client closed connection during header read. Underlying error",
					zap.String("address", conn.RemoteAddr().String()), zap.Error(headerReadError))
			}

			return
		}
		// Now turn that buffer of bytes into an integer - represnts size of message body
		msgLength, bytesParsed := byteArrayToUInt32(headerBuffer)
		iMsgLength := int(msgLength)
		// Not sure what the correct way to handle these errors are. For now, bomb out
		if bytesParsed == 0 {
			// "Buffer too small"
			zzlog.Errorw("0 Bytes parsed from header. Underlying error",
				zap.String("address", conn.RemoteAddr().String()), zap.Error(headerReadError))

			return
		} else if bytesParsed < 0 {
			// "Buffer overflow"
			zzlog.Errorw("Buffer Less than zero bytes parsed from header. Underlying error",
				zap.String("address", conn.RemoteAddr().String()), zap.Error(headerReadError))

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
				zzlog.Errorw("Failure to read from connection. ",
					zap.String("address", conn.RemoteAddr().String()),
					zap.Int64("msgLength", msgLength),
					zap.Int("totalDataBytesRead", totalDataBytesRead),
					zap.Error(dataReadError))
			} else {
				// The client wrote the header but closed the connection
				zzlog.Errorw("Client closed connection during data read. Underlying error",
					zap.String("address", conn.RemoteAddr().String()), zap.Error(dataReadError))
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
					zzlog.Errorw("Socket recv.Callback error", zap.Error(err))
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
func ReadFromTcp(conn *net.TCPConn, rcb recv, ccb closed) (err error) {
	handleListenedConn(conn, 4, DefaultMaxMessageSize, rcb, ccb)

	return
}

func WriteToConnections(conn *net.TCPConn, packet []byte) (n int, err error) {
	if 0 == len(packet) {
		return
	}

	if nil == conn {
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
		bytesWritten, writeError = conn.Write(toWrite[totalBytesWritten:])
		totalBytesWritten += bytesWritten
	}

	err = writeError
	n = totalBytesWritten
	return
}
