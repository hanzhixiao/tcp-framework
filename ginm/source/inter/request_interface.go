package inter

type HandleStep int
type Request interface {
	GetConn() Conn
	GetMessage() Message
	GetWorkerID() int
	GetMessageType() uint32
	GoTo(HandleStep)
	Abort()
	BindRouter(Router)
	Call()
	GetData() []byte
	RouterSlicesNext()
	BindRouterSlices(handlers []RouterHandler)
	BindWorker(workID int)
	SetResponse(response IcResp)
	GetResponse() IcResp
}

type BaseRequest struct{}

func (br *BaseRequest) SetResponse(response IcResp) {
	return
}

func (br *BaseRequest) GetResponse() IcResp {
	return nil
}

func (br *BaseRequest) BindRouterSlices(handlers []RouterHandler) {
	return
}
func (br *BaseRequest) BindWorker(workID int) {
	return
}

func (br *BaseRequest) BindRouter(Router) {
	return
}
func (br *BaseRequest) GetWorkerID() int {
	return 0
}

func (br *BaseRequest) GetMessageType() uint32 {
	return 0
}

func (br *BaseRequest) GetConn() Conn       { return nil }
func (br *BaseRequest) GetData() []byte     { return nil }
func (br *BaseRequest) GetMsgID() uint32    { return 0 }
func (br *BaseRequest) GetMessage() Message { return nil }
func (br *BaseRequest) Call()               {}
func (br *BaseRequest) Abort()              {}
func (br *BaseRequest) GoTo(HandleStep)     {}
func (br *BaseRequest) RouterSlicesNext()   {}
