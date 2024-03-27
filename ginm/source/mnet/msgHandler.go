package mnet

import (
	"context"
	"fmt"
	"math/rand"
	"mmo/ginm/pkg/common/config"
	"mmo/ginm/source/inter"
	"mmo/ginm/zlog"
)

type msgHandler struct {
	apis         map[uint32]inter.Router
	globalQueue  chan inter.Request
	workers      []inter.Worker
	workerNum    int
	ctx          context.Context
	cancel       context.CancelFunc
	routerSlices *routerSlices
	builder      *chainBuilder
	frameDecoder inter.Decoder
}

func (m *msgHandler) Exec(request inter.Request) {
	m.builder.Execute(request)
}

func (m *msgHandler) AddInterceptor(interceptor inter.Interceptor) {
	m.builder.AddInterceptor(interceptor)
}

func (m *msgHandler) AddRouterSlices(msgID uint32, router ...inter.RouterHandler) inter.RouterSlices {
	m.routerSlices.AddHandler(msgID, router...)
	return m.routerSlices
}

func (m *msgHandler) Use(handlers ...inter.RouterHandler) inter.RouterSlices {
	m.routerSlices.Use(handlers...)
	return m.routerSlices
}

func (m *msgHandler) Group(start uint32, end uint32, handlers ...inter.RouterHandler) inter.GroupRouterSlices {
	return NewGroup(start, end, m.routerSlices, handlers...)
}

func (m *msgHandler) StopAllWorkers() {
	m.cancel()
}

func (m *msgHandler) MsgToTaskQueue(request inter.Request) {
	m.GetWorker(request.GetWorkerID()).GetRequestQueue() <- request
	fmt.Println("request to msg queue ", request.GetWorkerID(), "successfully...")
}

func (m *msgHandler) assignWorker() int {
	workerId := rand.Intn(m.GetWorkerNum())
	return workerId
}

func (m *msgHandler) GetWorkerNum() int {
	return m.workerNum
}

func (m *msgHandler) GetWorker(workerId int) inter.Worker {
	return m.workers[workerId]
}

func (m *msgHandler) StartAllWorker(chanSize int) {
	for i := 0; i < len(m.workers); i++ {
		m.workers[i] = NewWorker(chanSize, i)
		fmt.Println("Worker ", i, " starts...")
		go m.StartOneWorker(m.workers[i])
	}
}

func (m *msgHandler) StartOneWorker(worker inter.Worker) {
	conf := config.GetConfig()
	for {
		select {
		case request := <-worker.GetRequestQueue():
			if !conf.GlobalObject.RouterSlicesMode {
				m.doMsgHandler(request)
			} else {
				m.doMsgHandlerSlices(request)
			}
		case <-m.ctx.Done():
			fmt.Println("Worker ", worker.GetWorkerId(), " stopped...")
			return
		default:

		}
	}
}

//func (m *msgHandler) robTask(worker inter.Worker) int {
//	for i := 0; i < 4; i++ {
//
//	}
//}

func (m *msgHandler) AddRouter(msgType uint32, router inter.Router) {
	if _, ok := m.apis[msgType]; ok {
		panic(fmt.Sprintf("Router %d has already existed", msgType))
	}
	m.apis[msgType] = router
	return
}

func (m *msgHandler) doMsgHandler(request inter.Request) {
	defer func() {
		if err := recover(); err != nil {
			zlog.Errorf("workerID: %d doMsgHandler panic: %v", request.GetWorkerID(), err)
		}
	}()
	msgType := request.GetMessageType()
	handler, ok := m.apis[msgType]
	_, ok2 := (request).(*RequestFunc)
	if !ok && !ok2 {
		zlog.Errorf("api msgID = %d is not FOUND!", request.GetMessageType())
		return
	}
	request.BindRouter(handler)
	request.Call()
}

func (m *msgHandler) doMsgHandlerSlices(request inter.Request) {
	defer func() {
		if err := recover(); err != nil {
			zlog.Errorf("workerID: %d doMsgHandler panic: %v", request.GetWorkerID(), err)
		}
	}()
	handlers, ok := m.routerSlices.GetHandlers(request.GetMessageType())
	if !ok {
		zlog.Errorf("api msgID = %d is not FOUND!", request.GetMessageType())
		return
	}
	request.BindRouterSlices(handlers)
	request.RouterSlicesNext()
}

func NewMessageHandler(workerNum int) inter.MsgHandler {
	mh := &msgHandler{apis: map[uint32]inter.Router{}, builder: newChainBuilder(), workers: make([]inter.Worker, workerNum), routerSlices: NewRouterSlices(), workerNum: workerNum}
	mh.ctx, mh.cancel = context.WithCancel(context.Background())
	mh.builder.Tail(mh)
	return mh
}

func (m *msgHandler) Intercept(chain inter.Chain) inter.IcResp {
	request := chain.Request()
	iRequest := request.(inter.Request)
	iRequest.BindWorker(m.assignWorker())
	m.workers[iRequest.GetWorkerID()].GetRequestQueue() <- iRequest
	return chain.Proceed(chain)
}
