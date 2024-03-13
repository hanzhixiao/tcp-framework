package main

import (
	"mmo/cmd/interceptor/interceptors"
	"mmo/cmd/interceptor/router"
	"mmo/ginm/source/mnet"
)

func main() {
	server := mnet.NewServer()

	server.AddRouter(1, &router.HelloRouter{})

	// Add Custom Interceptor
	server.AddInterceptor(&interceptors.MyInterceptor{})

	server.Serve()
}
