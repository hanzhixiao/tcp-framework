package mnet

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/xtaci/kcp-go"
	"mmo/ginm/pkg/common/config"
	"mmo/ginm/source/inter"
	"mmo/ginm/source/interceptor"
	"mmo/ginm/zlog"
	"net"
	"sync"
	"time"
)

type kcpConn struct {
	conn             *kcp.UDPSession
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
	closeLock        sync.Mutex
	frameDecoder     inter.FrameDecoder
	server           inter.Server
}

func (k *kcpConn) GetMsgHandler() inter.MsgHandler {
	return k.msgHandler
}

func (k *kcpConn) GetTCPConn() *net.TCPConn {
	return nil
}

func (k *kcpConn) GetConnID() uint32 {
	return k.connID
}

func (k *kcpConn) GetWsConn() *websocket.Conn {
	return nil
}

func (k *kcpConn) GetRemoteAddr() net.Addr {
	return k.conn.RemoteAddr()
}

func (k *kcpConn) Send(msgType uint32, buf []byte) error {
	dataLen := len(buf)
	msg := NewMessage(buf, uint32(dataLen))
	msg.SetMsgID(msgType)
	packData, err := msg.Pack()
	if err != nil {
		fmt.Println("Pack msg err: ", err.Error())
		return err
	}
	k.msgChan <- packData
	return nil
}

func (k *kcpConn) Listen() {
	fmt.Println("conn listening")
	go k.startReader()
	go k.startWriter()
	k.callOnConnStart()
}

func (k *kcpConn) Stop() {
	k.closeLock.Lock()
	if k.IsClosed() {
		fmt.Println("conn has already closed")
		return
	}
	k.setClose()
	k.closeLock.Unlock()
	k.callOnConnStop()
	if k.hc != nil {
		k.hc.Stop()
	}
	k.conn.Close()
	close(k.closedCh)
	close(k.msgChan)
	fmt.Println("tcp conn ", k.connID, " closed")
}

func (k *kcpConn) SetProperty(key string, value interface{}) {
	k.propertyLock.Lock()
	defer k.propertyLock.Unlock()

	k.property[key] = value
}

func (k *kcpConn) GetProperty(key string) (interface{}, error) {
	k.propertyLock.Lock()
	defer k.propertyLock.Unlock()

	if value, ok := k.property[key]; ok {
		return value, nil
	}

	return nil, errors.New("no property found")
}

func (k *kcpConn) RemoveProperty(key string) {
	k.propertyLock.Lock()
	defer k.propertyLock.Unlock()

	delete(k.property, key)
}

func (k *kcpConn) SetHeartBeat(checker inter.HeartbeatChecker) {
	k.hc = checker
}

func (k *kcpConn) Start() {
	if k.server.GetDecoder() != nil {
		k.frameDecoder = interceptor.NewFrameDecoder(k.server.GetDecoder().GetLengthField())
	}
	if k.hc != nil {
		k.hc.Start()
		k.updateActivity()
	}

	k.Listen()
}

func (c *kcpConn) startReader() {
	defer func() {
		c.Stop()
		fmt.Println("reader closed with err")
	}()
	for {
		select {
		case <-c.closedCh:
			return
		default:
			buff := make([]byte, 1024)
			n, err := c.conn.Read(buff)
			if err != nil {
				zlog.Errorf("read msg head [read datalen=%d], error = %s", n, err)
				return
			}
			if c.hc != nil {
				c.updateActivity()
			}
			fmt.Println("Server read data successfully:", buff[:n])
			if c.frameDecoder != nil {
				bufArrays := c.frameDecoder.Decode(buff[:n])
				if bufArrays == nil {
					continue
				}
				for _, buff := range bufArrays {
					msg := NewMessage(buff, uint32(len(buff)))
					request := NewRequest(msg, c)
					c.msgHandler.Exec(request)
				}
			} else {
				msg := NewMessage(buff[:n], uint32(n))
				request := NewRequest(msg, c)
				c.msgHandler.Exec(request)
			}
		}
	}
}

func (c *kcpConn) callOnConnStop() {
	if c.onConnStop != nil {
		zlog.Infof("ZINX CallOnConnStop....")
		c.onConnStop(c)
	}
}

func (c *kcpConn) callOnConnStart() {
	if c.onConnStart != nil {
		zlog.Infof("ZINX CallOnConnStart....")
		c.onConnStart(c)
	}
}

func (k *kcpConn) IsAlive() bool {
	conf := config.GetConfig()
	if k.isClosed {
		return false
	}
	return time.Now().Sub(k.lastActivityTime) < conf.HeartbeatMaxDuration()
}

func (k *kcpConn) LocalAddr() net.Addr {
	return k.conn.LocalAddr()
}
func (c *kcpConn) startWriter() {
	defer func() {
		c.Stop()
		fmt.Println("writer closed with err")
	}()
	fmt.Println("Writer goroutine for ", c.remoteAddr, " starts")
	for {
		select {
		case msg := <-c.msgChan:
			if _, err := c.conn.Write(msg); err != nil {
				fmt.Println("Server writes tcp err:", err.Error())
				return
			}
		case <-c.closedCh:
			return
		}
	}
}

func (k *kcpConn) updateActivity() {
	k.lastActivityTime = time.Now()
}

func (k *kcpConn) IsClosed() bool {
	return k.isClosed
}

func (k *kcpConn) setClose() {
	k.isClosed = true
}

func NewKcpServerConn(server inter.Server, conn *kcp.UDPSession, cid uint32, onConnStart, onConnStop inter.Hook, msgHandler inter.MsgHandler) inter.Conn {
	return &kcpConn{
		server:           server,
		conn:             conn,
		connID:           cid,
		lastActivityTime: time.Time{},
		localAddr:        conn.LocalAddr().String(),
		remoteAddr:       conn.RemoteAddr().String(),
		onConnStart:      onConnStart,
		onConnStop:       onConnStop,
		property:         make(map[string]interface{}),
		propertyLock:     sync.Mutex{},
		isClosed:         false,
		msgChan:          make(chan []byte, 1024),
		closedCh:         make(chan interface{}),
		closeLock:        sync.Mutex{},
		msgHandler:       msgHandler,
	}
}
