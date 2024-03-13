package mnet

import (
	"github.com/pkg/errors"
	"mmo/ginm/pkg/utils"
	"mmo/ginm/source/inter"
	"sync"
)

type connManager struct {
	connPools map[uint32]inter.Conn
	sync.RWMutex
}

func (c *connManager) ClearConn() {
	c.RWMutex.Lock()
	for connId, conn := range c.connPools {
		conn.Stop()
		delete(c.connPools, connId)
	}
	c.RWMutex.Unlock()
}

func (c *connManager) GetConn(connId uint32) (inter.Conn, error) {
	c.RWMutex.Lock()
	conn, ok := c.connPools[connId]
	c.RWMutex.Unlock()
	if !ok {
		return nil, utils.Wrap(errors.New("Get conn err"), "Conn not found")
	}
	return conn, nil
}

func (c *connManager) AddConn(conn inter.Conn) {
	c.RWMutex.Lock()
	c.connPools[conn.GetConnID()] = conn
	c.RWMutex.Unlock()
}

func (c *connManager) RemoveConn(connId uint32) error {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	_, ok := c.connPools[connId]
	if !ok {
		return utils.Wrap(errors.New("Remove Conn err"), "Conn not found")
	}
	delete(c.connPools, connId)
	return nil
}

func (c *connManager) GetConnNum() int {
	return len(c.connPools)
}

func NewConnManager() inter.ConnManager {
	return &connManager{connPools: make(map[uint32]inter.Conn)}
}
