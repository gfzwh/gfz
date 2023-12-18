package socket

type SocketOption func(*options)

type options struct {
	maxMessageSize int32
	headerByteSize int32
	enableLogging  bool
	address        string
	listen         listen
	connect        connect
	recv           recv
	closed         closed
}

func MaxMessageSize(maxMessageSize int32) SocketOption {
	return func(c *options) {
		c.maxMessageSize = maxMessageSize
	}
}

func HeaderByteSize(headerByteSize int32) SocketOption {
	return func(c *options) {
		c.headerByteSize = headerByteSize
	}
}

func EnableLogging(enableLogging bool) SocketOption {
	return func(c *options) {
		c.enableLogging = enableLogging
	}
}

func Address(address string) SocketOption {
	return func(c *options) {
		c.address = address
	}
}

func Listen(listen listen) SocketOption {
	return func(c *options) {
		c.listen = listen
	}
}

func Connect(connect connect) SocketOption {
	return func(c *options) {
		c.connect = connect
	}
}

func Recv(recv recv) SocketOption {
	return func(c *options) {
		c.recv = recv
	}
}

func Closed(closed closed) SocketOption {
	return func(c *options) {
		c.closed = closed
	}
}

func initOpts(opts ...SocketOption) *options {
	var opt options
	for _, o := range opts {
		o(&opt)
	}

	return &opt
}
