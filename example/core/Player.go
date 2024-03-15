package core

import (
	"encoding/json"
	"log"
	"math/rand"
	"mmo/example/Person"
	"mmo/example/util"
	"mmo/ginm/source/inter"
	"strconv"
	"sync"
)

type Json1 struct {
	Id int32 // 玩家id，数据库主键生成
}

type Json2 struct {
	X string
	Y string
}

type Json3 struct {
	Id       int32
	X        string
	Y        string
	State    Person.State
	MoveVecX string
	MoveVecY string
}

type Player struct {
	Conn  inter.Conn // 玩家连接
	Json1 *Json1
	Json2 *Json2
	Json3 *Json3
}

func (p *Player) SendMsg(msgId uint32, data []byte) {
	err := p.Conn.Send(msgId, data)
	if err != nil {
		return
	}
}

func (p *Player) SyncPid() {
	log.Println("SyncPid")
	data, err := json.Marshal(p.Json1) // TODO json
	if err != nil {
		log.Println(err)
		return
	}
	p.SendMsg(1, data)
}

func (p *Player) SyncPos() {
	log.Println("SyncPos")
	data, err := json.Marshal(p.Json2) // TODO json
	if err != nil {
		log.Println(err)
		return
	}
	p.SendMsg(2, data)
}

func (p *Player) SyncOtherPos(player *Player) {
	//log.Println("SyncOtherPos")
	data, err := json.Marshal(player.Json3)
	if err != nil {
		log.Println(err)
		return
	}
	p.SendMsg(3, data)
}

func (p *Player) SyncUnPid(pid int32) {
	log.Println("SyncUnPid")
	j := Json1{Id: pid}
	data, err := json.Marshal(j) // TODO json
	if err != nil {
		log.Println(err)
		return
	}
	p.SendMsg(4, data)
}

//////////////////////////////////////////////////////////////////////

var pidGen int32 = 0
var pidLock sync.Mutex

// float32 转 String工具类，保留2位小数
func FloatToString(input_num float32) string {
	return strconv.FormatFloat(float64(input_num), 'f', 2, 64)
}

func NewPlayer(conn inter.Conn) *Player {
	var vec util.Vector2

	pidLock.Lock()
	id := pidGen
	pidGen++
	vec.X = -4 + (rand.Float32()-0.5)*2
	vec.Y = 2 + (rand.Float32()-0.5)*2
	pidLock.Unlock()

	return &Player{
		Conn:  conn,
		Json1: &Json1{Id: id},
		Json2: &Json2{
			X: FloatToString(vec.X),
			Y: FloatToString(vec.Y),
		},
		Json3: &Json3{
			Id:       id,
			X:        FloatToString(vec.X),
			Y:        FloatToString(vec.Y),
			State:    0,
			MoveVecX: "0",
			MoveVecY: "0",
		},
	}
}
