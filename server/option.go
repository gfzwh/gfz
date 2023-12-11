package server

import "context"

type Options func(*options)

type options struct {
	bind string
	port int

	ctx context.Context
}

func Bind(addr string) Options {
	return func(c *options) {
		c.bind = addr
	}
}
func Port(port int) Options {
	return func(c *options) {
		c.port = port
	}
}

func SetOption(k, v interface{}) Options {
	return func(o *options) {
		if o.ctx == nil {
			o.ctx = context.Background()
		}
		o.ctx = context.WithValue(o.ctx, k, v)
	}
}
