package mnet

import (
	"bytes"
	"fmt"
	"mmo/ginm/source/inter"
	"mmo/ginm/zlog"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	StackBegin = 3
	StackEnd   = 5
)

type BaseRouter struct {
}

func (b *BaseRouter) PreHandler(request inter.Request) {
}

func (b *BaseRouter) Handler(request inter.Request) {

}

func (b *BaseRouter) PostHandler(request inter.Request) {

}

type routerSlices struct {
	apis     map[uint32][]inter.RouterHandler
	handlers []inter.RouterHandler
	sync.RWMutex
}

func (s *routerSlices) Use(handlers ...inter.RouterHandler) {
	s.handlers = append(s.handlers, handlers...)
}

func (s *routerSlices) AddHandler(msgID uint32, handlers ...inter.RouterHandler) {
	if _, ok := s.apis[msgID]; ok {
		panic("repeated api , msgId = " + strconv.Itoa(int(msgID)))
	}
	finalSize := len(s.handlers) + len(handlers)
	mergedHandlers := make([]inter.RouterHandler, finalSize)
	copy(mergedHandlers, s.handlers)
	copy(mergedHandlers[len(s.handlers):], handlers)
	s.apis[msgID] = append(s.apis[msgID], mergedHandlers...)
}

func (s *routerSlices) GetHandlers(msgId uint32) ([]inter.RouterHandler, bool) {
	handlers, ok := s.apis[msgId]
	return handlers, ok
}

func NewRouterSlices() *routerSlices {
	return &routerSlices{
		apis:     make(map[uint32][]inter.RouterHandler, 10),
		handlers: make([]inter.RouterHandler, 0, 6),
	}
}

type GroupRouter struct {
	start    uint32
	end      uint32
	handlers []inter.RouterHandler
	router   *routerSlices
}

func NewGroup(start, end uint32, router *routerSlices, Handlers ...inter.RouterHandler) *GroupRouter {
	g := &GroupRouter{
		start:    start,
		end:      end,
		handlers: make([]inter.RouterHandler, 0, len(Handlers)),
		router:   router,
	}
	g.handlers = append(g.handlers, Handlers...)
	return g
}

func (g *GroupRouter) AddHandler(msgID uint32, handlers ...inter.RouterHandler) {
	if msgID > g.end || msgID < g.start {
		panic("add router to group err in msgId:" + strconv.Itoa(int(msgID)))
	}
	finalSize := len(g.handlers) + len(handlers)
	mergedHandlers := make([]inter.RouterHandler, finalSize)
	copy(mergedHandlers, g.handlers)
	copy(mergedHandlers[len(g.handlers):], handlers)

	g.router.AddHandler(msgID, mergedHandlers...)
}

func (g *GroupRouter) Use(handlers ...inter.RouterHandler) {
	g.handlers = append(g.handlers, handlers...)
}

func RouterRecover(request inter.Request) {
	defer func() {
		if err := recover(); err != nil {
			panicInfo := getInfo(StackBegin)
			zlog.Errorf("MsgId:%d Handler panic: info:%s err:%v", request.GetMessageType(), panicInfo, err)
		}
	}()
	request.RouterSlicesNext()
}

func getInfo(ship int) (infoStr string) {
	panicInfo := new(bytes.Buffer)
	for i := ship; i <= StackEnd; i++ {
		pc, file, lineNo, ok := runtime.Caller(i)
		if !ok {
			break
		}
		funcname := runtime.FuncForPC(pc).Name()
		filename := path.Base(file)
		funcname = strings.Split(funcname, ".")[1]
		fmt.Fprintf(panicInfo, "funcname:%s filename:%s LineNo:%d\n", funcname, filename, lineNo)
	}
	return panicInfo.String()
}

func RouterTime(request inter.Request) {
	now := time.Now()
	request.RouterSlicesNext()
	duration := time.Since(now)
	fmt.Println(duration.String())
}
