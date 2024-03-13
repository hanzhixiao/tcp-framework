package decoder

import (
	"encoding/hex"
	"math"
	"mmo/ginm/source/inter"
	"mmo/ginm/zlog"
)

const HEADER_SIZE = 5

type HtlvCrcDecoder struct {
	Head    byte
	Funcode byte
	Length  byte
	Body    []byte
	Crc     []byte
	Data    []byte
}

func (h *HtlvCrcDecoder) Intercept(chain inter.Chain) inter.IcResp {
	message := chain.GetMessage()
	if message == nil {
		return chain.ProceedWithIMessage(message, nil)
	}
	data := message.GetData()
	if len(data) < HEADER_SIZE {
		return chain.ProceedWithIMessage(message, nil)
	}
	htlvData := h.decode(data)
	message.SetMsgID(uint32(htlvData.Funcode))
	return chain.ProceedWithIMessage(message, *htlvData)
}

func (h *HtlvCrcDecoder) decode(data []byte) *HtlvCrcDecoder {
	dataSize := len(data)
	h.Data = data
	h.Head = data[0]
	h.Funcode = data[1]
	h.Length = data[2]
	h.Body = data[3 : dataSize-2]
	h.Crc = data[dataSize-2:]
	if !checkCRC(data[:dataSize-2], h.Crc) {
		zlog.Debugf("crc check error %s %s\n", hex.EncodeToString(data), hex.EncodeToString(h.Crc))
		return nil
	}
	return h
}

func NewHTLVCRCDecoder() inter.Decoder {
	return &HtlvCrcDecoder{}
}

func (h HtlvCrcDecoder) GetLengthField() *inter.LengthField {
	return &inter.LengthField{
		MaxFrameLength:      math.MaxInt8 + 4,
		LengthFieldOffset:   2,
		LengthFieldLength:   1,
		LengthAdjustment:    2,
		InitialBytesToStrip: 0,
	}
}
