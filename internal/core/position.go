package core

type Position struct {
	x float32
	y float32
	z float32
	v float32
}

func NewPosition(x float32, y float32, z float32, v float32) *Position {
	return &Position{x: x, y: y, z: z, v: v}
}
