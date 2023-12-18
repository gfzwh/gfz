package zzlog

type LoggerOption func(*options)

type options struct {
	level   string
	logName string
}

func WithLevel(level string) LoggerOption {
	return func(args *options) {
		args.level = level
	}
}

func WithLogName(logName string) LoggerOption {
	return func(args *options) {
		args.logName = logName
	}
}

func initOpts(opts ...LoggerOption) *options {
	var opt options
	for _, o := range opts {
		o(&opt)
	}

	return &opt
}
