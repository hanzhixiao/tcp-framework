package inter

import "time"

type Hook func(conn Conn)

type Server interface {
	Start(port ...int)
	Serve(port ...int)
	Stop()
	AddRouter(msgType uint32, router Router)
	SetOnStartConn(Hook)
	SetOnStopConn(Hook)
	GetOnStartConn() Hook
	GetOnStopConn() Hook
	StartHeartBeat(time.Duration)
	Group(start, end uint32, Handler ...RouterHandler) GroupRouterSlices
	Use(handlers ...RouterHandler) RouterSlices
	AddRouterSlices(msgID uint32, router ...RouterHandler) RouterSlices
	AddInterceptor(interceptor Interceptor)
	SetDecoder(Decoder)
	GetFieldLength() *LengthField
	GetDecoder() Decoder
}
