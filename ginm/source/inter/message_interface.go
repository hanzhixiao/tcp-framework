package inter

import (
	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
	"net"
)

type Message interface {
	PackHtlvCrc() ([]byte, error)
	Pack() ([]byte, error)
	Unpack(conn net.Conn, wsConn *websocket.Conn, session *kcp.UDPSession) error
	GetData() []byte
	GetDataLen() uint32
	GetHeaderLen() uint32
	GetMsgType() uint32
	SetMsgID(msgID uint32)
	SetData(value []byte)
	SetDataLen(length uint32)
}
