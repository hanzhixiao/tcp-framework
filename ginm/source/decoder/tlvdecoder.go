package decoder

import (
	"bytes"
	"encoding/binary"
	"math"
	"mmo/ginm/source/inter"
)

const TLV_HEADER_SIZE = 8 //表示TLV空包长度

type TLVDecoder struct {
	Tag    uint32 //T
	Length uint32 //L
	Value  []byte //V
}

func NewTLVDecoder() inter.Decoder {
	return &TLVDecoder{}
}

func (tlv *TLVDecoder) GetLengthField() *inter.LengthField {
	return &inter.LengthField{
		MaxFrameLength:      math.MaxUint32 + 4 + 4,
		LengthFieldOffset:   0,
		LengthFieldLength:   4,
		LengthAdjustment:    4,
		InitialBytesToStrip: 0,
	}
}

func (tlv *TLVDecoder) decode(data []byte) *TLVDecoder {
	tlvData := TLVDecoder{}
	tlvData.Length = binary.LittleEndian.Uint32(data[0:4])
	tlvData.Tag = binary.LittleEndian.Uint32(data[4:8])
	tlvData.Value = make([]byte, tlvData.Length)

	binary.Read(bytes.NewBuffer(data[8:8+tlvData.Length]), binary.LittleEndian, tlvData.Value)

	return &tlvData
}

func (tlv *TLVDecoder) Intercept(chain inter.Chain) inter.IcResp {
	iMessage := chain.GetMessage()
	if iMessage == nil {
		return chain.ProceedWithIMessage(iMessage, nil)
	}

	data := iMessage.GetData()
	if len(data) < TLV_HEADER_SIZE {
		return chain.ProceedWithIMessage(iMessage, nil)
	}

	tlvData := tlv.decode(data)
	iMessage.SetMsgID(tlvData.Tag)
	iMessage.SetData(tlvData.Value)
	iMessage.SetDataLen(tlvData.Length)

	return chain.ProceedWithIMessage(iMessage, *tlvData)
}
