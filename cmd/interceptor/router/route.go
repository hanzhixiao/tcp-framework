package router

import (
	"mmo/ginm/source/inter"
	"mmo/ginm/source/mnet"
	"mmo/ginm/zlog"
)

type HelloRouter struct {
	mnet.BaseRouter
}

func (hr *HelloRouter) Handler(request inter.Request) {
	zlog.Infof(string(request.GetData()))
}
