package inter

import "encoding/binary"

type LengthField struct {
	Order               binary.ByteOrder
	MaxFrameLength      int64
	LengthFieldOffset   int
	LengthFieldLength   int
	LengthAdjustment    int
	InitialBytesToStrip int
}
