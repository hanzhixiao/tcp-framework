package inter

type Decoder interface {
	Interceptor
	GetLengthField() *LengthField
}

type FrameDecoder interface {
	Decode(buff []byte) [][]byte
}
