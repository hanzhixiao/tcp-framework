# tcp-framework
#### 快速开始
1. 创建server服务实例
2. 配置自定义路由及业务
3. 启动服务

```golang
package main

import (
	"fmt"
	"mmo/ginm/pkg/utils"
	"mmo/ginm/source/inter"
	"mmo/ginm/source/mnet"
	"time"
)
type pingRouter struct {
	mnet.BaseRouter
}
func (r *pingRouter) Handler(request inter.Request) {
	data := append(request.GetMessage().GetData(), utils.StringtoSlice("Handler\n")...)
	if err := request.GetConn().Send(request.GetMessage().GetMsgType(), data); err != nil {
		fmt.Println("Handler error:", err.Error())
		return
	}
}

func main() {
	tcpServer := mnet.NewServer()
	tcpServer.AddRouter(1, &pingRouter{})
	tcpServer.StartHeartBeat(time.Second)
	tcpServer.Serve()
}
 ```

###示例程序
为了

