package main

import (
	"mmo/cmd/async_op/router"
	"mmo/ginm/source/inter"
	"mmo/ginm/source/mnet"
	"mmo/ginm/zlog"
)

func OnConnectionAdd(conn inter.Conn) {
	zlog.Debug("async_op OnConnectionAdd ===>")
}

func OnConnectionLost(conn inter.Conn) {
	zlog.Debug("async_op OnConnectionLost ===>")
}

func main() {
	s := mnet.NewServer()

	s.SetOnStartConn(OnConnectionAdd)
	s.SetOnStopConn(OnConnectionLost)

	s.AddRouter(1, &router.LoginRouter{})

	s.Serve()
}
