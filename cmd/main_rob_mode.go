package main

import (
	"fmt"
	"mmo/ginm/source/inter"
	"mmo/ginm/source/mnet"
	"time"
)

type Router1 struct {
	mnet.BaseRouter
}
type Router2 struct {
	mnet.BaseRouter
}

func (r *Router2) Handler(request inter.Request) {
	if err := request.GetConn().Send(1, []byte("pong")); err != nil {
		fmt.Println("Handler error:", err.Error())
		return
	}
}

func (r *Router1) Handler(request inter.Request) {
	time.Sleep(5 * time.Second)
	if err := request.GetConn().Send(2, []byte("pong10")); err != nil {
		fmt.Println("Handler error:", err.Error())
		return
	}
}

func main() {
	tcpServer := mnet.NewServer()
	tcpServer.AddRouter(1, &Router1{})
	tcpServer.AddRouter(2, &Router2{})
	tcpServer.Serve()
}
