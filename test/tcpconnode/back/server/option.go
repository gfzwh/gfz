package server

import "context"

type HandlerOption func(*options)

type options struct {
	bind string
	port int

	ctx context.Context
}

func Bind(addr string) HandlerOption {
	return func(c *options) {
		c.bind = addr
	}
}
func Port(port int) HandlerOption {
	return func(c *options) {
		c.port = port
	}
}

func SetOption(k, v interface{}) HandlerOption {
	return func(o *options) {
		if o.ctx == nil {
			o.ctx = context.Background()
		}
		o.ctx = context.WithValue(o.ctx, k, v)
	}
}
