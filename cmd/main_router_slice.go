package main

import (
	"fmt"
	"mmo/ginm/source/inter"
	"mmo/ginm/source/mnet"
)

func Test1(request inter.Request) {
	fmt.Println("test1")
}
func Test2(request inter.Request) {
	fmt.Println("Test2")
}
func Test3(request inter.Request) {
	fmt.Println("Test3")
}
func Test4(request inter.Request) {
	fmt.Println("Test4")
}
func Test5(request inter.Request) {
	fmt.Println("Test5")
}
func Test6(request inter.Request) {
	fmt.Println("Test6")
}

type router struct {
	mnet.BaseRouter
}

func (r *router) PreHandle(req inter.Request) {
	fmt.Println(" hello router1")
}
func (r *router) Handle(req inter.Request) {
	req.Abort()
	fmt.Println(" hello router2")
}
func (r *router) PostHandle(req inter.Request) {
	fmt.Println(" hello router3")
}

func main() {

	{
		server := mnet.NewServer()
		group := server.Group(3, 10, Test1)

		group.AddHandler(3, Test2)

		server.AddRouterSlices(1, Test3)

		group.Use(Test2, Test3)
		group.AddHandler(5, Test4, Test5, Test6)

		router := server.Use(Test4, Test5)
		router.AddHandler(2, Test6)

		group.AddHandler(4, Test6)

		server.Serve()
	}

}
