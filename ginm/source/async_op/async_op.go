package async_op

import "sync"

var asyncWorkerArray = [2048]*AsyncWorker{}
var initAsyncWorkerWorkerLocker = &sync.Mutex{}

func Process(opID int, asyncOp func()) {
	if asyncOp == nil {
		return
	}
	curWork := getCurWorker(opID)
	if curWork != nil {
		curWork.process(asyncOp)
	}
}

func getCurWorker(opID int) *AsyncWorker {
	if opID < 0 {
		opID = -opID
	}

	workIndex := opID % len(asyncWorkerArray)
	curWorker := asyncWorkerArray[workIndex]
	if curWorker != nil {
		return curWorker
	}
	initAsyncWorkerWorkerLocker.Lock()
	defer initAsyncWorkerWorkerLocker.Unlock()
	curWorker = asyncWorkerArray[workIndex]
	if curWorker != nil {
		return curWorker
	}
	curWorker = &AsyncWorker{taskQue: make(chan func(), 2048)}
	go curWorker.loopExecTask()
	return curWorker
}
