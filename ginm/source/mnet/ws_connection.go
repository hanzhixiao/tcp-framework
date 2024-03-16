package mnet

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"mmo/ginm/pkg/common/config"
	"mmo/ginm/source/inter"
	"mmo/ginm/source/interceptor"
	"mmo/ginm/zlog"
	"net"
	"sync"
	"time"
)

type wsConnection struct {
	conn             *websocket.Conn
	server           inter.Server
	connID           uint32
	lastActivityTime time.Time
	localAddr        string
	remoteAddr       string
	hc               inter.HeartbeatChecker
	onConnStart      inter.Hook
	onConnStop       inter.Hook
	property         map[string]interface{}
	propertyLock     sync.Mutex
	isClosed         bool
	msgHandler       inter.MsgHandler
	msgChan          chan []byte
	closedCh         chan interface{}
	frameDecoder     inter.FrameDecoder
}

func (w *wsConnection) GetMsgChan() chan []byte {
	return w.msgChan
}

func (w *wsConnection) GetMsgHandler() inter.MsgHandler {
	return w.msgHandler
}

func (w *wsConnection) GetWsConn() *websocket.Conn {
	return w.conn
}

func (w *wsConnection) GetTCPConn() *net.TCPConn {
	return nil
}

func (w *wsConnection) GetConnID() uint32 {
	return w.connID
}

func (w *wsConnection) GetRemoteAddr() net.Addr {
	return w.conn.RemoteAddr()
}

func (w *wsConnection) Send(msgType uint32, buf []byte) error {
	dataLen := len(buf)
	msg := NewMessage(buf, uint32(dataLen))
	msg.SetMsgID(msgType)
	packData, err := msg.Pack()
	if err != nil {
		fmt.Println("Pack msg err: ", err.Error())
		return err
	}
	w.msgChan <- packData
	return nil
}

func (w *wsConnection) Listen() {
	fmt.Println("conn listening")
	go w.startReader()
	go w.startWriter()
	w.callOnConnStart()
}

func (c *wsConnection) Stop() {
	if c.IsClosed() {
		fmt.Println("conn has already closed")
		return
	}
	c.setClose()
	c.callOnStopConn()
	if c.hc != nil {
		c.hc.Stop()
	}
	c.conn.Close()
	close(c.closedCh)
	close(c.msgChan)
	fmt.Println("tcp conn ", c.connID, " closed")
}

func (w *wsConnection) SetProperty(key string, value interface{}) {
	w.propertyLock.Lock()
	defer w.propertyLock.Unlock()
	w.property[key] = value
}

func (w *wsConnection) GetProperty(key string) (interface{}, error) {
	w.propertyLock.Lock()
	defer w.propertyLock.Unlock()
	property, ok := w.property[key]
	if !ok {
		return nil, errors.New(fmt.Sprintf("property %s not exist", property))
	}
	return property, nil
}

func (w *wsConnection) RemoveProperty(key string) {
	w.propertyLock.Lock()
	defer w.propertyLock.Unlock()
	delete(w.property, key)
}

func (w *wsConnection) SetHeartBeat(checker inter.HeartbeatChecker) {
	w.hc = checker
}

func (c *wsConnection) Start() {
	if c.server.GetDecoder() != nil {
		c.frameDecoder = interceptor.NewFrameDecoder(c.server.GetDecoder().GetLengthField())
	}
	if c.hc != nil {
		c.hc.Start()
		c.updateActivity()
	}

	// Start the Goroutine for users to read data from the client.
	// (开启用户从客户端读取数据流程的Goroutine)
	c.Listen()
}
func (c *wsConnection) callOnConnStart() {
	if c.onConnStart != nil {
		zlog.Infof("ZINX CallOnConnStart....")
		c.onConnStart(c)
	}
}

func (c *wsConnection) updateActivity() {
	c.lastActivityTime = time.Now()
}

func (c *wsConnection) startReader() {
	defer func() {
		c.Stop()
		fmt.Println("reader closed with err")
	}()
	for {
		select {
		case <-c.closedCh:
			return
		default:
			_, p, err := c.conn.ReadMessage()
			if err != nil {
				zlog.Errorf("read message from websocket conn error: %s", err.Error())
				return
			}
			request := NewRequest(NewMessage(p, uint32(len(p))), c)
			if c.hc != nil {
				c.updateActivity()
			}
			fmt.Println("Server read data successfully:", string(request.GetMessage().GetData()))
			c.msgHandler.Exec(request)
		}
	}
}

func (w *wsConnection) IsAlive() bool {
	if w.IsClosed() {
		return false
	}
	conf := config.GetConfig()
	return time.Now().Sub(w.lastActivityTime) < conf.HeartbeatMaxDuration()
}

func (w *wsConnection) LocalAddr() net.Addr {
	return w.conn.LocalAddr()
}

func (w *wsConnection) IsClosed() bool {
	return w.isClosed
}

func (w *wsConnection) setClose() {
	w.isClosed = true
}

func (w *wsConnection) callOnStopConn() {
	w.onConnStop(w)
}

func (w *wsConnection) startWriter() {
	defer func() {
		w.Stop()
		fmt.Println("writer closed with err")
	}()
	fmt.Println("Writer goroutine for ", w.remoteAddr, " starts")
	for {
		select {
		case msg := <-w.msgChan:
			if err := w.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				fmt.Println("Server writes tcp err:", err.Error())
				return
			}
		case <-w.closedCh:
			return
		}
	}
}

func NewWsConnection(server inter.Server, conn *websocket.Conn, connID uint32, msgHandler inter.MsgHandler, onConnStart, onConnStop inter.Hook) inter.Conn {
	return &wsConnection{
		server:      server,
		conn:        conn,
		connID:      connID,
		isClosed:    false,
		msgChan:     nil,
		property:    nil,
		localAddr:   conn.LocalAddr().String(),
		remoteAddr:  conn.RemoteAddr().String(),
		msgHandler:  msgHandler,
		onConnStart: onConnStart,
		onConnStop:  onConnStop,
	}
}
