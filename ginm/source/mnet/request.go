package mnet

import (
	"mmo/ginm/pkg/common/config"
	"mmo/ginm/source/inter"
	"sync"
)

const (
	PRE_HANDLE  inter.HandleStep = iota // PreHandle for pre-processing
	HANDLE                              // Handle for processing
	POST_HANDLE                         // PostHandle for post-processing

	HANDLE_OVER
)

type request struct {
	inter.BaseRequest
	message  inter.Message
	router   inter.Router
	conn     inter.Conn
	stepLock *sync.Mutex
	steps    inter.HandleStep
	needNext bool
	handlers []inter.RouterHandler
	index    int8 //路由处理函数切片索引
	workerID int
	icResp   inter.IcResp
}

func (r *request) BindRouterSlices(handlers []inter.RouterHandler) {
	r.handlers = handlers
}

func NewRequest(message inter.Message, conn inter.Conn) inter.Request {
	return &request{message: message, stepLock: new(sync.Mutex), conn: conn, steps: PRE_HANDLE, needNext: true, index: -1}
}

func (r *request) GetConn() inter.Conn {
	return r.conn
}

func (r *request) GetMessage() inter.Message {
	return r.message
}

func (r *request) GetMessageType() uint32 {
	return r.message.GetMsgType()
}

func (r *request) GoTo(step inter.HandleStep) {
	r.stepLock.Lock()
	defer r.stepLock.Unlock()
	r.steps = step
	r.needNext = false
}
func (r *request) Abort() {
	conf := config.GetConfig()
	if conf.GlobalObject.RouterSlicesMode {
		r.index = int8(len(r.handlers))
	} else {
		r.stepLock.Lock()
		r.steps = HANDLE_OVER
		r.stepLock.Unlock()
	}
}

func (r *request) BindRouter(router inter.Router) {
	r.router = router
}

func (r *request) Call() {
	if r.router == nil {
		return
	}
	for r.steps < HANDLE_OVER {
		switch r.steps {
		case PRE_HANDLE:
			r.router.PreHandler(r)
		case HANDLE:
			r.router.Handler(r)
		case POST_HANDLE:
			r.router.PostHandler(r)
		}
		r.next()
	}
	r.steps = PRE_HANDLE
}

func (r *request) GetResponse() inter.IcResp {
	return r.icResp
}

func (r *request) SetResponse(response inter.IcResp) {
	r.icResp = response
}

func (r *request) next() {
	if r.needNext == false {
		r.needNext = true
		return
	}
	r.stepLock.Lock()
	r.steps++
	r.stepLock.Unlock()
}

func (r *request) GetData() []byte {
	return r.message.GetData()
}

func (r *request) RouterSlicesNext() {
	r.index++
	for r.index < int8(len(r.handlers)) {
		r.handlers[r.index](r)
		r.index++
	}
}

func (r *request) BindWorker(workerID int) {
	r.workerID = workerID
}
func (r *request) GetWorkerID() int {
	return r.workerID
}

type RequestFunc struct {
	inter.BaseRequest
	conn     inter.Conn
	callFunc func()
	workerID int
}

func (rf *RequestFunc) GetConnection() inter.Conn {
	return rf.conn
}

func (rf *RequestFunc) GetWorkerID() int {
	return rf.workerID
}

func (rf *RequestFunc) Call() {
	if rf.callFunc != nil {
		rf.callFunc()
	}
}

func NewFuncRequest(conn inter.Conn, callFunc func()) inter.Request {
	req := new(RequestFunc)
	req.conn = conn
	req.callFunc = callFunc
	return req
}
