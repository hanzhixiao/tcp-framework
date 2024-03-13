package internal

type Grid struct {
	minX   int
	maxX   int
	minY   int
	maxY   int
	player map[int]bool
}
