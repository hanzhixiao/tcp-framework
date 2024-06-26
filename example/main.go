package main

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"math"
	"mmo/example/Person"
	"mmo/example/core"
	"mmo/ginm/source/decoder"
	"mmo/ginm/source/inter"
	"mmo/ginm/source/mnet"
)

func onConnStart(conn inter.Conn) {
	log.Println(conn.GetConnID(), "上线")
	player := core.NewPlayer(conn)
	player.SyncPid()
	player.SyncPos()

	players := wm.GetAllPlayers()
	for _, play := range players {
		if play.Json1.Id == player.Json1.Id {
			continue
		}
		log.Println("SyncOtherPos in onConnStart")
		play.SyncOtherPos(player)
		player.SyncOtherPos(play)
	}
	wm.Add(player)
}

func onConnStop(conn inter.Conn) {
	log.Println(conn.GetConnID(), "下线")
	id := int32(conn.GetConnID())
	wm.Remove(id)

	players := wm.GetAllPlayers()
	for _, play := range players {
		play.SyncUnPid(id)
	}
}

type moveRouter struct {
	mnet.BaseRouter
}

func (pr *moveRouter) Handler(request inter.Request) {
	player := wm.GetPlayerById(int32(request.GetConn().GetConnID()))
	msg := request.GetResponse()
	coder := msg.(decoder.TLVDecoder)
	data := coder.Value
	if err := json.Unmarshal(data, player.Json3); err != nil {
		log.Println(err.Error())
		return
	}

	players := wm.GetAllPlayers()
	if player.Json3.State == Person.Attack {
		log.Println(player.Json1.Id, " Attack")
	}
	for _, play := range players {
		if play.Json1.Id == player.Json1.Id {
			continue
		}
		play.SyncOtherPos(player)
	}
}

var wm *core.WorldManager

type coder struct {
	Funcode uint32
	Length  uint32
	Body    []byte
}

func (c *coder) GetLengthField() *inter.LengthField {
	return &inter.LengthField{
		MaxFrameLength:      math.MaxInt8 + 4,
		LengthFieldOffset:   0,
		LengthFieldLength:   4,
		LengthAdjustment:    0,
		InitialBytesToStrip: 0,
	}
}

func (c *coder) Intercept(chain inter.Chain) inter.IcResp {
	request := chain.Request()
	iRequest := request.(inter.Request)
	data := iRequest.GetMessage().GetData()
	c.Body = data[8:]
	c.Funcode = binary.BigEndian.Uint32(data[4:8])
	c.Length = binary.BigEndian.Uint32(data[:4])
	return c
}

func main() {
	s := mnet.NewServer()
	wm = core.NewWorldManager()

	pr := &moveRouter{}
	s.AddRouter(201, pr) // 201 移动请求
	//s.SetDecoder(decoder.NewHTLVCRCDecoder())
	s.SetOnStartConn(onConnStart)
	s.SetOnStopConn(onConnStop)
	s.Serve()
}
