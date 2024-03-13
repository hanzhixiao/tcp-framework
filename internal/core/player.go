package core

import (
	"errors"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"mmo/ginm/pkg/utils"
	"mmo/ginm/source/inter"
	"mmo/pb/pb_message"
	"mmo/pkg/constant"
	"sync"
)

type Player struct {
	pid      int32
	conn     inter.Conn
	position *Position
}

const offsetX = 10
const offsetY = 10

var pidGen int32 = 1
var lock sync.Mutex

func NewPlayer(conn inter.Conn) *Player {
	player := &Player{conn: conn, position: NewPosition(float32(160+rand.Intn(offsetX)), 0, float32(160+rand.Intn(offsetY)), 0)}
	lock.Lock()
	defer lock.Unlock()
	player.pid = pidGen
	pidGen++
	return player
}

func (p *Player) SendMessage(msgId uint32, data proto.Message) error {
	marData, err := proto.Marshal(data)
	if err != nil {
		return utils.Wrap(errors.New("SendMessage err: "), "marshal data failed")
	}
	if err := p.conn.Send(msgId, marData); err != nil {
		return err
	}
	return nil
}

func (p *Player) SyncPid() error {
	data := pb_message.SyncPid{Pid: p.pid}
	if err := p.SendMessage(constant.LoginMsg, &data); err != nil {
		return err
	}
	return nil
}

func (p *Player) BroadCastStartPosition() error {
	data := pb_message.BroadCast{Pid: p.pid, Tp: constant.Position, Data: &pb_message.BroadCast_P{
		&pb_message.Position{
			X: p.position.x,
			Y: p.position.y,
			Z: p.position.z,
			V: p.position.v,
		},
	}}
	if err := p.SendMessage(constant.BroadcastMsg, &data); err != nil {
		return err
	}
	return nil
}
