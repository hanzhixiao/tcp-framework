package inter

type Router interface {
	PreHandler(request Request)
	Handler(request Request)
	PostHandler(request Request)
}

type RouterHandler func(request Request)

type GroupRouterSlices interface {
	AddHandler(msgID uint32, handlers ...RouterHandler)
	Use(handlers ...RouterHandler)
}

type RouterSlices interface {
	Use(Handlers ...RouterHandler)
	AddHandler(msgId uint32, handlers ...RouterHandler)
	GetHandlers(MsgId uint32) ([]RouterHandler, bool)
}
