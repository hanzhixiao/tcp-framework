package mnet

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"mmo/ginm/pkg/common/config"
	"mmo/ginm/pkg/utils"
	"mmo/ginm/source/inter"
	"mmo/ginm/source/interceptor"
	"mmo/ginm/zlog"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type conn struct {
	server           inter.Server
	tcpConn          *net.TCPConn
	connId           uint32
	isClosed         int32
	msgChan          chan []byte
	msgHandler       inter.MsgHandler
	closedCh         chan struct{}
	property         map[string]interface{}
	propertyLock     sync.RWMutex
	hc               inter.HeartbeatChecker
	lastActivityTime time.Time
	frameDecoder     inter.FrameDecoder
}

func (c *conn) GetMsgChan() chan []byte {
	return c.msgChan
}
func (c *conn) GetMsgHandler() inter.MsgHandler {
	return c.msgHandler
}

func (c *conn) GetWsConn() *websocket.Conn {
	return nil
}

func (c *conn) LocalAddr() net.Addr {
	return c.tcpConn.LocalAddr()
}

func (c *conn) IsAlive() bool {
	if c.IsClosed() {
		return false
	}
	conf := config.GetConfig()
	return time.Now().Sub(c.lastActivityTime) < conf.HeartbeatMaxDuration()
}

func (c *conn) Start() {
	if c.server.GetDecoder() != nil {
		c.frameDecoder = interceptor.NewFrameDecoder(c.server.GetDecoder().GetLengthField())
	}
	if c.hc != nil {
		c.hc.Start()
		c.updateActivity()
	}
	c.Listen()
}

func (c *conn) SetHeartBeat(checker inter.HeartbeatChecker) {
	c.hc = checker
}

func (c *conn) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	c.property[key] = value
}

func (c *conn) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	val, ok := c.property[key]
	if !ok {
		fmt.Println("unExisted key ", key)
		return nil, utils.Wrap(errors.New("unExisted key"), key)
	}
	return val, nil
}

func (c *conn) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	delete(c.property, key)
}

func (c *conn) callOnStartConn() {
	onStartConn := c.server.GetOnStartConn()
	if onStartConn == nil {
		fmt.Println("onStartConn Hook is not set")
		return
	}
	onStartConn(c)
}

func (c *conn) callOnStopConn() {
	onStopConn := c.server.GetOnStopConn()
	if onStopConn == nil {
		fmt.Println("onStopConn Hook is not set")
		return
	}
	onStopConn(c)
}

func (c *conn) GetTCPConn() *net.TCPConn {
	return c.tcpConn
}

func (c *conn) GetConnID() uint32 {
	return c.connId
}

func (c *conn) GetRemoteAddr() net.Addr {
	return c.tcpConn.RemoteAddr()
}

func (c *conn) Send(msgType uint32, buf []byte) error {
	dataLen := len(buf)
	msg := NewMessage(buf, uint32(dataLen))
	msg.SetMsgID(msgType)
	packData, err := msg.Pack()
	if err != nil {
		fmt.Println("Pack msg err: ", err.Error())
		return err
	}
	select {
	case c.msgChan <- packData:
	case <-c.closedCh:
		return nil
	}
	return nil
}

func (c *conn) reader() {
	defer func() {
		c.Stop()
		fmt.Println("reader closed with err")
	}()
	defer func() {
		if err := recover(); err != nil {
			zlog.Ins().ErrorF("connID=%d, panic err=%v", c.GetConnID(), err)
		}
	}()
	for {
		select {
		case <-c.closedCh:
			return
		default:
			buff := make([]byte, 1024)
			n, err := c.tcpConn.Read(buff)
			if err != nil {
				zlog.Errorf("read msg head [read datalen=%d], error = %s", n, err)
				return
			}
			if c.hc != nil {
				c.updateActivity()
			}
			fmt.Println("Server read data successfully:", string(buff[:n]))
			if c.frameDecoder != nil {
				bufArrays := c.frameDecoder.Decode(buff[:n])
				if bufArrays == nil {
					continue
				}
				for _, buff := range bufArrays {
					fmt.Println(string(buff))
					msg := NewMessage(buff, uint32(len(buff)))
					request := NewRequest(msg, c)
					c.msgHandler.Exec(request)
				}
			} else {
				msg := NewMessage(buff[0:n], uint32(n))
				request := NewRequest(msg, c)
				c.msgHandler.Exec(request)
			}
		}
	}
}
func (c *conn) writer() {
	defer func() {
		c.Stop()
		fmt.Println("writer closed with err")
	}()
	fmt.Println("Writer goroutine for ", c.tcpConn.RemoteAddr().String(), " starts")
	for {
		select {
		default:
			msg := <-c.msgChan
			if _, err := c.tcpConn.Write(msg); err != nil {
				fmt.Println("Server writes tcp err:", err.Error())
				return
			}
		case <-c.closedCh:
			return
		}
	}
}
func (c *conn) Listen() {
	fmt.Println("conn listening")
	go c.reader()
	go c.writer()
	c.callOnStartConn()
}

func (c *conn) Stop() {
	if c.IsClosed() {
		fmt.Println("conn has already closed")
		return
	}
	if !c.setClose() {
		return
	}
	c.callOnStopConn()
	if c.hc != nil {
		c.hc.Stop()
	}
	c.tcpConn.Close()
	close(c.closedCh)
	close(c.msgChan)
	fmt.Println("tcp conn ", c.connId, " closed")
}

func (c *conn) setClose() bool {
	return atomic.CompareAndSwapInt32(&c.isClosed, 0, 1)
}

func (c *conn) updateActivity() {
	c.lastActivityTime = time.Now()
}

func (c *conn) IsClosed() bool {
	return atomic.LoadInt32(&c.isClosed) != 0
}

func NewConn(server inter.Server, tcpConn *net.TCPConn, connId uint32, msgHandler inter.MsgHandler) inter.Conn {
	c := &conn{
		server:     server,
		tcpConn:    tcpConn,
		connId:     connId,
		isClosed:   0,
		msgHandler: msgHandler,
		closedCh:   make(chan struct{}),
		msgChan:    make(chan []byte, 1024),
		property:   make(map[string]interface{}),
	}
	fieldLength := server.GetFieldLength()
	if fieldLength != nil {
		c.frameDecoder = interceptor.NewFrameDecoder(fieldLength)
	}
	return c
}
