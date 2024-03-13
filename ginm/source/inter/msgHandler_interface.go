package inter

type MsgHandler interface {
	AddRouter(msgType uint32, router Router)
	StartAllWorker(chanSize int)
	GetWorker(workerId int) Worker
	GetWorkerNum() int
	MsgToTaskQueue(request Request)
	StopAllWorkers()
	Group(start uint32, end uint32, handler ...RouterHandler) GroupRouterSlices
	Use(handlers ...RouterHandler) RouterSlices
	AddRouterSlices(msgID uint32, router ...RouterHandler) RouterSlices
	AddInterceptor(interceptor Interceptor)
	Exec(request Request)
}
