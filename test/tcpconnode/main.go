package main

import (
	"os"
	"os/signal"
	"syscall"
	"tcpconnode/controller"
	"tcpconnode/protocol"

	"github.com/shockerjue/gfz/server"
	"github.com/shockerjue/gfz/zzlog"
)

func main() {
	svr := server.NewServer()
	ctl := controller.Controller()

	err := protocol.RegisterUserServiceHandler(svr, ctl)
	if nil != err {
		zzlog.Fatalf("RegisterUserServiceHandler error, errinfo:%v\n", err)

		return
	}
	svr.Run(server.Bind("127.0.0.1"), server.Port(8989))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	return
}
