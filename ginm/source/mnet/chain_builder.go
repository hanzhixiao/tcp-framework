package mnet

import "mmo/ginm/source/inter"

type chainBuilder struct {
	body       []inter.Interceptor
	head, tail inter.Interceptor
}

func (c *chainBuilder) AddInterceptor(interceptor inter.Interceptor) {
	c.body = append(c.body, interceptor)
}

func newChainBuilder() *chainBuilder {
	return &chainBuilder{
		body: make([]inter.Interceptor, 0),
	}
}

func (c *chainBuilder) Head(interceptor inter.Interceptor) {
	c.head = interceptor
}

func (c *chainBuilder) Tail(interceptor inter.Interceptor) {
	c.tail = interceptor
}

func (c *chainBuilder) Execute(request inter.IcReq) inter.IcResp {
	var interceptors []inter.Interceptor
	if c.head != nil {
		interceptors = append(interceptors, c.head)
	}
	if len(c.body) > 0 {
		interceptors = append(interceptors, c.body...)
	}
	if c.tail != nil {
		interceptors = append(interceptors, c.tail)
	}

	// Create a new interceptor chain and execute each interceptor
	chain := NewChain(request, 0, interceptors)

	// Execute the chain
	return chain.Proceed(request)
}

type chain struct {
	req          inter.IcReq
	position     int
	interceptors []inter.Interceptor
}

func (c *chain) ProceedWithIMessage(message inter.Message, response inter.IcReq) inter.IcResp {
	if message == nil || response == nil {
		return c.Proceed(c.Request())
	}
	req := c.Request()
	if req == nil {
		return c.Proceed(c.Request())
	}
	iRequest := c.ShouldRequest(req)
	if iRequest == nil {
		return c.Proceed(c.Request())
	}
	iRequest.SetResponse(response)
	return c.Proceed(iRequest)
}

func (c *chain) GetMessage() inter.Message {
	req := c.Request()
	if req == nil {
		return nil
	}
	iReq := c.ShouldRequest(req)
	if iReq == nil {
		return nil
	}
	return iReq.GetMessage()
}

func NewChain(req inter.IcReq, position int, interceptors []inter.Interceptor) inter.Chain {
	return &chain{req: req, position: position, interceptors: interceptors}
}

func (c *chain) Request() inter.IcReq {
	return c.req
}

func (c *chain) Proceed(req inter.IcReq) inter.IcResp {
	if c.position < len(c.interceptors) {
		interceptor := c.interceptors[c.position]
		c.position++
		response := interceptor.Intercept(c)
		return response
	}
	return req
}

func (c *chain) ShouldRequest(req inter.IcReq) inter.Request {
	if req == nil {
		return nil
	}
	switch req.(type) {
	case inter.Request:
		return req.(inter.Request)
	default:
		return nil
	}
}
