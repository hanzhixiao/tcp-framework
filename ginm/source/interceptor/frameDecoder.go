package interceptor

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"mmo/ginm/source/inter"
	"sync"
)

type frameDecoder struct {
	inter.LengthField
	LengthFieldEndOffset int
	in                   []byte
	lock                 sync.Mutex
	discardTooLongFrame  bool
	bytesToDiscard       int64
	tooLongFrameLength   int64
}

func NewFrameDecoder(lf *inter.LengthField) inter.FrameDecoder {
	decoder := &frameDecoder{
		LengthField:          *lf,
		in:                   make([]byte, 0),
		lock:                 sync.Mutex{},
		LengthFieldEndOffset: lf.LengthFieldOffset + lf.LengthFieldLength,
	}
	if decoder.Order == nil {
		decoder.Order = binary.LittleEndian
	}
	return decoder
}

func (f *frameDecoder) Decode(buff []byte) [][]byte {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.in = append(f.in, buff...)
	resp := make([][]byte, 0)
	for {
		arr := f.decode(f.in)
		if arr != nil {
			resp = append(resp, arr)
			_size := len(arr) + f.InitialBytesToStrip
			if _size > 0 {
				f.in = f.in[_size:]
			}
		} else {
			return resp
		}
	}
}

func (f *frameDecoder) decode(in []byte) []byte {
	buf := bytes.NewBuffer(in)
	if f.discardTooLongFrame {
		f.discardTooLongFrameFunc(buf)
	}
	if buf.Len() < f.LengthFieldEndOffset {
		return nil
	}
	frameLength := f.getUnadjustedFrameLength(buf, f.LengthFieldOffset, f.LengthFieldLength, f.Order)
	if frameLength < 0 {
		f.failOnNegativeLengthField(buf, frameLength, f.LengthFieldEndOffset)
	}
	frameLength += int64(f.LengthFieldEndOffset) + int64(f.LengthAdjustment)
	if frameLength > f.MaxFrameLength {
		f.exceededFrameLength(buf, frameLength)
		return nil
	}
	if buf.Len() < int(frameLength) {
		return nil
	}
	if f.InitialBytesToStrip > int(frameLength) {
		f.failOnFrameLengthLessThanInitialBytesToStrip(buf, frameLength, f.InitialBytesToStrip)
	}
	buf.Next(f.InitialBytesToStrip)
	actualFrameLength := int(frameLength) - f.InitialBytesToStrip
	buff := make([]byte, actualFrameLength)
	buf.Read(buff)
	//bytes.NewBuffer([]byte{})
	//_in := bytes.NewBuffer(buff)
	return buff
}

func (f *frameDecoder) exceededFrameLength(buf *bytes.Buffer, frameLength int64) {
	discard := frameLength - int64(buf.Len())
	f.tooLongFrameLength = frameLength
	if discard < 0 {
		buf.Next(int(frameLength))
	} else {
		f.bytesToDiscard = discard
		f.discardTooLongFrame = true
		buf.Next(buf.Len())
	}
	f.failIfNecessary()
}

func (d *frameDecoder) failOnNegativeLengthField(in *bytes.Buffer, frameLength int64, lengthFieldEndOffset int) {
	in.Next(lengthFieldEndOffset)
	panic(fmt.Sprintf("negative pre-adjustment length field: %d", frameLength))
}

func (f *frameDecoder) discardTooLongFrameFunc(buffer *bytes.Buffer) {
	bytesToDiscard := f.bytesToDiscard
	localBytesToDiscard := math.Min(float64(bytesToDiscard), float64(buffer.Len()))
	buffer.Next(int(localBytesToDiscard))
	f.bytesToDiscard -= int64(localBytesToDiscard)
	f.failIfNecessary()
}

func (f *frameDecoder) failIfNecessary() {
	if f.bytesToDiscard == 0 {
		f.discardTooLongFrame = false
	}
}

func (f *frameDecoder) getUnadjustedFrameLength(buf *bytes.Buffer, offset int, length int, order binary.ByteOrder) int64 {
	arr := buf.Bytes()
	arr = arr[offset : offset+length]
	buffer := bytes.NewBuffer(arr)
	var frameLength int64
	switch length {
	case 1:
		var value int8
		binary.Read(buffer, order, &value)
		frameLength = int64(value)
	case 2:
		var value uint16
		binary.Read(buffer, order, &value)
		frameLength = int64(value)
	case 3:
		//int占32位，这里取出后24位，返回int类型
		if order == binary.LittleEndian {
			n := uint(arr[0]) | uint(arr[1])<<8 | uint(arr[2])<<16
			frameLength = int64(n)
		} else {
			n := uint(arr[2]) | uint(arr[1])<<8 | uint(arr[0])<<16
			frameLength = int64(n)
		}
	case 4:
		//int
		var value uint32
		binary.Read(buffer, order, &value)
		frameLength = int64(value)
	case 8:
		//long
		binary.Read(buffer, order, &frameLength)
	default:
		panic(fmt.Sprintf("unsupported LengthFieldLength: %d (expected: 1, 2, 3, 4, or 8)", f.LengthFieldLength))
	}
	return frameLength
}

func (f *frameDecoder) failOnFrameLengthLessThanInitialBytesToStrip(in *bytes.Buffer, frameLength int64, initialBytesToStrip int) {
	in.Next(int(frameLength))
	panic(fmt.Sprintf("Adjusted frame length (%d) is less  than InitialBytesToStrip: %d", frameLength, initialBytesToStrip))
}
