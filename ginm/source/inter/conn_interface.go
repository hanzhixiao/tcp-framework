package inter

import (
	"github.com/gorilla/websocket"
	"net"
)

type Conn interface {
	GetMsgChan() chan []byte
	GetTCPConn() *net.TCPConn
	GetConnID() uint32
	GetWsConn() *websocket.Conn
	GetRemoteAddr() net.Addr
	GetMsgHandler() MsgHandler
	Send(msgId uint32, buf []byte) error
	Listen()
	Stop()
	SetProperty(key string, value interface{})
	GetProperty(key string) (interface{}, error)
	RemoveProperty(key string)
	SetHeartBeat(checker HeartbeatChecker)
	Start()
	IsAlive() bool
	LocalAddr() net.Addr
}
