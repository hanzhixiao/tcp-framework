package core

import (
	"fmt"
	"github.com/pkg/errors"
	"mmo/ginm/pkg/utils"
)

type GridManager struct {
	minX  int
	maxX  int
	cntX  int
	minY  int
	maxY  int
	cntY  int
	grids map[int]*Grid
}

func NewGridManager(minX int, maxX int, cntX int, minY int, maxY int, cntY int) *GridManager {
	manager := GridManager{minX: minX, maxX: maxX, cntX: cntX, minY: minY, maxY: maxY, cntY: cntY, grids: make(map[int]*Grid)}
	gridLength := manager.getGridLength()
	gridWidth := manager.getGridWidth()
	for i := 0; i < cntY; i++ {
		for j := 0; j < cntX; j++ {
			gridId := i*cntX + j
			manager.grids[gridId] = NewGrid(gridId, j*gridLength, (j+1)*gridLength, i*gridWidth, (i+1)*gridWidth)
		}
	}
	return &manager
}

func (m *GridManager) getGridLength() int {
	return (m.maxX - m.minX) / m.cntX
}

func (m *GridManager) getGridWidth() int {
	return (m.maxY - m.minY) / m.cntY
}

func (m *GridManager) GetSurroundedGridByGid(gridId int) ([]*Grid, error) {
	var surroundedGrids []*Grid
	grid, ok := m.grids[gridId]
	if !ok {
		return nil, utils.Wrap(errors.New("GetSurroundedGridByGid err: "), fmt.Sprintf("Grid %d doesn't exist", gridId))
	}
	surroundedGrids = append(surroundedGrids, grid)

	idx := m.GetIdx(gridId)
	if idx > 0 {
		surroundedGrids = append(surroundedGrids, m.grids[gridId-1])
	}
	if idx < m.cntX-1 {
		surroundedGrids = append(surroundedGrids, m.grids[gridId+1])
	}
	for _, surroundedGrid := range surroundedGrids {
		idy := m.GetIdy(surroundedGrid.gridId)
		if idy > 0 {
			surroundedGrids = append(surroundedGrids, m.grids[surroundedGrid.gridId-m.cntX])
		}
		if idy < m.cntY-1 {
			surroundedGrids = append(surroundedGrids, m.grids[surroundedGrid.gridId+m.cntX])
		}
	}
	return surroundedGrids, nil
}

func (m *GridManager) String() string {
	s := fmt.Sprintf("minX: %d, maxX: %d, cntX: %d, minY: %d, maxY: %d, cntY: %d", m.minX, m.maxX, m.cntX, m.minY, m.maxY, m.cntY)
	for _, grid := range m.grids {
		s = fmt.Sprintln(grid)
	}
	return s
}

func (m *GridManager) GetIdx(gid int) int {
	return gid % m.cntX
}

func (m *GridManager) GetIdy(gid int) int {
	return gid / m.cntX
}

func (m *GridManager) AddPlayerByGid(gid int, pid int) error {
	if err := m.grids[gid].AddPlayer(pid); err != nil {
		return err
	}
	return nil
}

func (m *GridManager) GetPlayerByGid(gid int) []int {
	return m.grids[gid].GetAllPlayers()
}

func (m *GridManager) RemovePlayerByGid(gid int, pid int) error {
	if err := m.grids[gid].RemovePlayer(pid); err != nil {
		return err
	}
	return nil
}

func (m *GridManager) GetGidByPos(x float32, y float32) int {
	return m.GetIdyByPos(y)*m.cntX + m.GetIdxByPos(x)
}

func (m *GridManager) GetIdxByPos(x float32) int {
	return int(x) - m.minX/m.getGridLength()
}

func (m *GridManager) GetIdyByPos(y float32) int {
	return int(y) - m.minY/m.getGridWidth()
}

func (m *GridManager) AddPlayerByPos(pid int, x float32, y float32) error {
	gid := m.GetGidByPos(x, y)
	return m.AddPlayerByGid(gid, pid)
}

func (m *GridManager) GetPlayerByPos(x float32, y float32) []int {
	gid := m.GetGidByPos(x, y)
	return m.GetPlayerByGid(gid)
}

func (m *GridManager) RemovePlayerByPos(x, y float32, pid int) error {
	gid := m.GetGidByPos(x, y)
	return m.RemovePlayerByGid(gid, pid)
}

func (m *GridManager) GetSurroundedPidsByGid(gid int) []int {
	return nil
}
