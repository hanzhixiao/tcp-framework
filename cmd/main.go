package main

import (
	"fmt"
	"mmo/ginm/pkg/utils"
	"mmo/ginm/source/inter"
	"mmo/ginm/source/mnet"
)

type pingRouter struct {
	mnet.BaseRouter
}

//func (r *pingRouter) PreHandler(request inter.Request) {
//	data := append(request.GetMessage().GetData(), utils.StringtoSlice("PreHandler\n")...)
//	if err := request.GetConn().Send(request.GetMessage().GetMsgType(), data); err != nil {
//		fmt.Println("PreHandler error:", err.Error())
//		return
//	}
//}

func (r *pingRouter) Handler(request inter.Request) {
	if err := request.GetConn().Send(1, []byte("pong")); err != nil {
		fmt.Println("Handler error:", err.Error())
		return
	}
}

//func (r *pingRouter) PostHandler(request inter.Request) {
//	data := append(request.GetMessage().GetData(), utils.StringtoSlice("PostHandler\n")...)
//	if err := request.GetConn().Send(request.GetMessage().GetMsgType(), data); err != nil {
//		fmt.Println("Post Handler error:", err.Error())
//		return
//	}
//}

type helloRouter struct {
	mnet.BaseRouter
}

func (r *helloRouter) PreHandler(request inter.Request) {
	data := append(request.GetMessage().GetData(), utils.StringtoSlice("PreHandler2\n")...)
	if err := request.GetConn().Send(request.GetMessage().GetMsgType(), data); err != nil {
		fmt.Println("PreHandler error:", err.Error())
		return
	}
}

func (r *helloRouter) Handler(request inter.Request) {
	data := append(request.GetMessage().GetData(), utils.StringtoSlice("Handler2\n")...)
	if err := request.GetConn().Send(request.GetMessage().GetMsgType(), data); err != nil {
		fmt.Println("Handler error:", err.Error())
		return
	}
}

func (r *helloRouter) PostHandler(request inter.Request) {
	data := append(request.GetMessage().GetData(), utils.StringtoSlice("PostHandler2\n")...)
	if err := request.GetConn().Send(request.GetMessage().GetMsgType(), data); err != nil {
		fmt.Println("Post Handler error:", err.Error())
		return
	}
}

func main() {
	tcpServer := mnet.NewServer()
	tcpServer.AddRouter(1, &pingRouter{})
	tcpServer.AddRouter(2, &helloRouter{})
	tcpServer.Serve()
}
