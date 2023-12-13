package client

import "context"

type CallOption func(*Options)

type Options struct {
	onlyCall bool
	timeout  int32

	ctx context.Context
}

func OnlyCall(onlyCall bool) CallOption {
	return func(args *Options) {
		args.onlyCall = onlyCall
	}
}

func Timeout(timeout int32) CallOption {
	return func(args *Options) {
		args.timeout = timeout
	}
}

func SetOption(k, v interface{}) CallOption {
	return func(o *Options) {
		if o.ctx == nil {
			o.ctx = context.Background()
		}
		o.ctx = context.WithValue(o.ctx, k, v)
	}
}

func initOpt(opts ...CallOption) *Options {
	var opt Options
	for _, o := range opts {
		o(&opt)
	}
	if 0 == opt.timeout {
		opt.timeout = 3
	}

	return &opt
}
