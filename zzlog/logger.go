package zzlog

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	defaultLogger *zap.SugaredLogger
)

func init() {
	logger, _ := zap.NewProduction()
	defaultLogger = logger.Sugar()

	return
}

func Init(opts ...LoggerOption) {
	opt := initOpts(opts...)
	if 0 == len(opt.logName) {
		return
	}

	var level zapcore.Level
	level, err := zapcore.ParseLevel(opt.level)
	if nil != err {
		level = zapcore.InfoLevel

	}

	lumberjacklogger := &lumberjack.Logger{
		Filename:   opt.logName,
		MaxSize:    5, // megabytes
		MaxBackups: 3,
		MaxAge:     7,    //days
		Compress:   true, // disabled by default
	}

	// defer lumberjacklogger.Close()
	multi := io.MultiWriter(os.Stdout, lumberjacklogger)

	// 创建一个新的 Logger
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder // 设置时间格式
	fileEncoder := zapcore.NewJSONEncoder(config)
	core := zapcore.NewCore(
		fileEncoder,            //编码设置
		zapcore.AddSync(multi), //输出到文件
		level,                  //日志等级
	)
	logger := zap.New(core, zap.AddCallerSkip(1), zap.AddCaller())
	// defer logger.Sync() // 在程序退出时确保日志缓冲区被刷新
	defaultLogger = logger.Sugar()
}

func DPanic(args ...interface{}) {
	defaultLogger.DPanic(args...)
}

func DPanicf(template string, args ...interface{}) {
	defaultLogger.DPanicf(template, args...)
}

func DPanicln(args ...interface{}) {
	defaultLogger.DPanicln(args...)
}

func DPanicw(msg string, keysAndValues ...interface{}) {
	defaultLogger.DPanicw(msg, keysAndValues...)
}

func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	defaultLogger.Debugf(template, args...)
}

func Debugln(args ...interface{}) {
	defaultLogger.Debugln(args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	defaultLogger.Debugw(msg, keysAndValues...)
}

func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	defaultLogger.Errorf(template, args...)
}

func Errorln(args ...interface{}) {
	defaultLogger.Errorln(args...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	defaultLogger.Errorw(msg, keysAndValues...)
}

func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	defaultLogger.Fatalf(template, args...)
}

func Fatalln(args ...interface{}) {
	defaultLogger.Fatalln(args...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	defaultLogger.Fatalw(msg, keysAndValues...)
}

func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	defaultLogger.Infof(template, args...)
}

func Infoln(args ...interface{}) {
	defaultLogger.Infoln(args...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	defaultLogger.Infow(msg, keysAndValues...)
}

func Panic(args ...interface{}) {
	defaultLogger.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	defaultLogger.Panicf(template, args...)
}

func Panicln(args ...interface{}) {
	defaultLogger.Panicln(args...)
}

func Panicw(msg string, keysAndValues ...interface{}) {
	defaultLogger.Panicw(msg, keysAndValues...)
}

func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	defaultLogger.Warnf(template, args...)
}

func Warnln(args ...interface{}) {
	defaultLogger.Warnln(args...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	defaultLogger.Warnw(msg, keysAndValues...)
}
