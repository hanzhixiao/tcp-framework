package interceptors

import (
	"mmo/ginm/source/inter"
	"mmo/ginm/zlog"
)

// Custom Interceptor 1

type MyInterceptor struct{}

func (m *MyInterceptor) Intercept(chain inter.Chain) inter.IcResp {
	request := chain.Request()
	// This layer is the custom interceptor processing logic, which simply prints the input.
	// (这一层是自定义拦截器处理逻辑，这里只是简单打印输入)
	iRequest := request.(inter.Request)
	zlog.Infof("MyInterceptor, Recv：%s", iRequest.GetData())
	return chain.Proceed(chain.Request())
}
