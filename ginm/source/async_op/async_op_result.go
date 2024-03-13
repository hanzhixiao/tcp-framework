package async_op

import (
	"mmo/ginm/source/inter"
	"mmo/ginm/source/mnet"
	"sync/atomic"
)

type AsyncOpResult struct {
	conn                             inter.Conn
	workerID                         int
	returnedObj                      interface{}
	completeFunc                     func()
	hasReturnObj                     int32
	hasCompleteFunc                  int32
	completeFuncHasAlreadyBeenCalled int32
}

func NewAsyncOpResult(conn inter.Conn, workerID int) *AsyncOpResult {
	return &AsyncOpResult{conn: conn, workerID: workerID}
}

func (a *AsyncOpResult) SetAsyncOpResult(val interface{}) {
	if atomic.CompareAndSwapInt32(&a.hasReturnObj, 0, 1) {
		a.returnedObj = val
		a.doComplete()
	}
}

func (a *AsyncOpResult) doComplete() {
	if a.completeFunc == nil {
		return
	}
	if atomic.CompareAndSwapInt32(&a.completeFuncHasAlreadyBeenCalled, 0, 1) {
		request := mnet.NewFuncRequest(a.conn, a.completeFunc)
		request.BindWorker(a.workerID)
		a.conn.GetMsgHandler().MsgToTaskQueue(request)
	}
}
func (a *AsyncOpResult) GetAsyncOpResult() interface{} {
	return a.returnedObj
}

func (a *AsyncOpResult) OnComplete(val func()) {
	if atomic.CompareAndSwapInt32(&a.hasCompleteFunc, 0, 1) {
		a.completeFunc = val
		if atomic.LoadInt32(&a.hasReturnObj) == 1 {
			a.doComplete()
		}
	}
}
