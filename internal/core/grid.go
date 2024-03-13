package core

import (
	"fmt"
	"github.com/pkg/errors"
	"mmo/ginm/pkg/utils"
	"sync"
)

type Grid struct {
	gridId     int
	minX       int
	maxX       int
	minY       int
	maxY       int
	players    map[int]bool
	playerLock sync.RWMutex
}

func NewGrid(gridId int, minX int, maxX int, minY int, maxY int) *Grid {
	return &Grid{gridId: gridId, minX: minX, maxX: maxX, minY: minY, maxY: maxY, players: make(map[int]bool)}
}

func (g *Grid) AddPlayer(playerId int) error {
	g.playerLock.Lock()
	defer g.playerLock.Unlock()
	if _, ok := g.players[playerId]; ok {
		return utils.Wrap(errors.New("AddPlayer err: "), fmt.Sprintf("player %d already exist in grid %d", playerId, g.gridId))
	}
	g.players[playerId] = true
	return nil
}

func (g *Grid) GetAllPlayers() []int {
	g.playerLock.RLock()
	defer g.playerLock.RUnlock()
	players := make([]int, len(g.players))
	for playerId, _ := range g.players {
		players = append(players, playerId)
	}
	return players
}

func (g *Grid) RemovePlayer(playerId int) error {
	g.playerLock.Lock()
	defer g.playerLock.Unlock()
	if _, ok := g.players[playerId]; !ok {
		return utils.Wrap(errors.New("RemovePlayer err: "), fmt.Sprintf("player %d doesn't exist in grid %d", playerId, g.gridId))
	}
	delete(g.players, playerId)
	return nil
}

func (g *Grid) String() string {
	return fmt.Sprintf("GridId: %d, players: %v, minX: %d, maxX: %d, minY: %d, maxY: %d", g.gridId, g.players, g.minX, g.maxX, g.minY, g.maxY)
}
