package inter

type HeartBeatMsgFunc func(conn Conn) []byte

type OnRemoteNotAlive func(Conn)
type HeartBeatFunc func(Conn) error

type HeartbeatChecker interface {
	SetOnRemoteNotAlive(OnRemoteNotAlive)
	SetHeartbeatMsgFunc(HeartBeatMsgFunc)
	SetHeartbeatFunc(HeartBeatFunc)
	MsgID() uint32
	Router() Router
	Clone() HeartbeatChecker
	BindConn(Conn)
	Start()
	Stop()
}

const (
	HeartBeatDefaultMsgID uint32 = 99999
)
