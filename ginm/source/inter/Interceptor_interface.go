package inter

type IcReq interface{}

type IcResp interface{}

type Chain interface {
	Request() IcReq
	Proceed(IcReq) IcResp
	GetMessage() Message
	ProceedWithIMessage(message Message, req IcReq) IcResp
}

type Interceptor interface {
	Intercept(Chain) IcResp
}
