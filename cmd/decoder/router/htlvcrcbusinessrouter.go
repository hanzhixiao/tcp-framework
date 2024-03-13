package router

import (
	"encoding/hex"
	"mmo/ginm/source/decoder"
	"mmo/ginm/source/inter"
	"mmo/ginm/source/mnet"
	"mmo/ginm/zlog"
)

type HtlvCrcBusinessRouter struct {
	mnet.BaseRouter
}

func (this *HtlvCrcBusinessRouter) Handler(request inter.Request) {

	//MsgID
	msgID := request.GetMessage().GetMsgType()
	zlog.Debugf("Call HtlvCrcBusinessRouter Handle %d %s\n", msgID, hex.EncodeToString(request.GetMessage().GetData()))

	resp := request.GetResponse()
	if resp == nil {
		return
	}

	tlvData := resp.(decoder.HtlvCrcDecoder)

	zlog.Debugf("do msgid=0x10 data business %+v\n", string(tlvData.Body))
}
