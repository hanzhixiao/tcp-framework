package main

import (
	"mmo/cmd/decoder/router"
	"mmo/ginm/source/decoder"
	"mmo/ginm/source/mnet"
)

func main() {
	s := mnet.NewServer()

	// TLV protocol corresponding to business function
	// TLV协议对应业务功能
	s.AddRouter(0x00000001, &router.HtlvCrcBusinessRouter{})

	// Process HTLVCRC protocol data
	// 处理HTLVCRC协议数据
	s.SetDecoder(decoder.NewHTLVCRCDecoder())

	// TLV protocol corresponding to business function, because the funcode field in client.go is 0x10
	// TLV协议对应业务功能，因为client.go中模拟数据funcode字段为0x10
	s.AddRouter(0x10, &router.HtlvCrcBusinessRouter{})

	// TLV protocol corresponding to business function, because the funcode field in client.go is 0x13
	// TLV协议对应业务功能，因为client.go中模拟数据funcode字段为0x13
	s.AddRouter(0x13, &router.HtlvCrcBusinessRouter{})

	//开启服务
	s.Serve()
}
