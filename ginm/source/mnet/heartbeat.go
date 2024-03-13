package mnet

import (
	"fmt"
	"mmo/ginm/source/inter"
	"mmo/ginm/zlog"
	"time"
)

type heartbeatChecker struct {
	interval         time.Duration
	quitChan         chan bool
	makeMsg          inter.HeartBeatMsgFunc
	onRemoteNotAlive inter.OnRemoteNotAlive
	msgID            uint32
	router           inter.Router
	routerSlices     []inter.RouterHandler
	conn             inter.Conn
	beatFunc         inter.HeartBeatFunc
}

func (h *heartbeatChecker) SetOnRemoteNotAlive(f inter.OnRemoteNotAlive) {
	if f != nil {
		h.onRemoteNotAlive = f
	}
}

func (h *heartbeatChecker) SetHeartbeatMsgFunc(f inter.HeartBeatMsgFunc) {
	if f != nil {
		h.makeMsg = f
	}
}

func (h *heartbeatChecker) SetHeartbeatFunc(beatFunc inter.HeartBeatFunc) {
	if beatFunc != nil {
		h.beatFunc = beatFunc
	}
}

func (h *heartbeatChecker) Stop() {
	zlog.Ins().InfoF("heartbeat checker stop, connID=%+v", h.conn.GetConnID())
	h.quitChan <- true
}

func (h *heartbeatChecker) Start() {
	go h.start()
}

func (h *heartbeatChecker) BindConn(conn inter.Conn) {
	h.conn = conn
	conn.SetHeartBeat(h)
}

func (h *heartbeatChecker) Clone() inter.HeartbeatChecker {
	heartbeat := &heartbeatChecker{
		interval:         h.interval,
		quitChan:         make(chan bool),
		beatFunc:         h.beatFunc,
		makeMsg:          h.makeMsg,
		onRemoteNotAlive: h.onRemoteNotAlive,
		msgID:            h.msgID,
		router:           h.router,
		routerSlices:     h.routerSlices,
		conn:             nil, // The bound connection needs to be reassigned
	}

	return heartbeat
}

type heartBeatDefaultRouter struct {
	BaseRouter
}

func (r *heartBeatDefaultRouter) Handle(req inter.Request) {
	zlog.Ins().InfoF("Recv Heartbeat from %s, MsgID = %+v, Data = %s",
		req.GetConn().GetRemoteAddr(), req.GetMessageType(), string(req.GetData()))
}

func NewHeartbeatChecker(interval time.Duration) inter.HeartbeatChecker {
	heartbeat := heartbeatChecker{
		interval:         interval,
		quitChan:         make(chan bool),
		makeMsg:          makeDefaultMsg,
		onRemoteNotAlive: notAliveDefaultFunc,
		msgID:            inter.HeartBeatDefaultMsgID,
		router:           &heartBeatDefaultRouter{},
		beatFunc:         nil,
	}
	return &heartbeat
}

func notAliveDefaultFunc(conn inter.Conn) {
	zlog.Infof("Remote connection %s is not alive, stop it", conn.GetRemoteAddr())
	conn.Stop()
}

func makeDefaultMsg(conn inter.Conn) []byte {
	msg := fmt.Sprintf("heartbeat [%s->%s]", conn.LocalAddr(), conn.GetRemoteAddr())
	return []byte(msg)
}

func (h *heartbeatChecker) MsgID() uint32 {
	return h.msgID
}

func (h *heartbeatChecker) Router() inter.Router {
	return h.router
}

func (h *heartbeatChecker) start() {
	ticker := time.NewTicker(h.interval)
	for {
		select {
		case <-ticker.C:
			h.check()
		case <-h.quitChan:
			ticker.Stop()
			return
		}
	}
}

func (h *heartbeatChecker) check() (err error) {
	if h.conn == nil {
		return nil
	}
	if !h.conn.IsAlive() {
		h.onRemoteNotAlive(h.conn)
	} else {
		if h.beatFunc != nil {
			err = h.beatFunc(h.conn)
		} else {
			err = h.SendHeartBeatMsg()
		}
	}
	return err
}

func (h *heartbeatChecker) SendHeartBeatMsg() error {
	msg := h.makeMsg(h.conn)
	if err := h.conn.Send(h.msgID, msg); err != nil {
		zlog.Error("send heartbeat msg error: %v, msgId=%+v msg=%+v", err, h.msgID, msg)
		return err
	}
	return nil
}
