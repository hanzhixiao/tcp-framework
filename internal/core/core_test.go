package core

import (
	"fmt"
	"testing"
)

func TestSurround(t *testing.T) {
	gridManager := NewGridManager(0, 250, 5, 0, 250, 5)
	for gid, _ := range gridManager.grids {
		fmt.Println("----------------------------------------------------------------", gid)
		fmt.Println(gridManager.GetSurroundedGridByGid(gid))
	}
}
