package test

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
	"mmo/ginm/source/inter"
	"mmo/ginm/source/mnet"
	"net"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestClient1(t *testing.T) {
	conn, err := net.Dial("tcp", "0.0.0.0:8099")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("client start successfully")
	for {
		message1 := mnet.NewMessage([]byte("Hello ginm"), 10)
		message1.SetMsgID(1)
		pack1, _ := message1.Pack()
		message2 := mnet.NewMessage([]byte("Hello ginm2"), 11)
		message2.SetMsgID(2)
		pack2, _ := message2.Pack()
		pack := append(pack1, pack2...)
		if _, err := conn.Write(pack); err != nil {
			fmt.Println("client write err:", err.Error())
			continue
		}
		fmt.Println("client write successfully")
		buf := make([]byte, 512)
		if _, err := conn.Read(buf); err != nil {
			fmt.Println("client read err", err.Error())
			continue
		}
		fmt.Println("client read successfully:", string(buf))
		time.Sleep(1 * time.Second)
	}
	runtime.Goexit()
}

func TestClient2(t *testing.T) {
	conn, err := net.Dial("tcp", "0.0.0.0:8099")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("client start successfully")
	for {
		message1 := mnet.NewMessage([]byte("2Hello ginm"), 11)
		message1.SetMsgID(1)
		pack1, _ := message1.Pack()
		message2 := mnet.NewMessage([]byte("2Hello ginm2"), 12)
		message2.SetMsgID(2)
		pack2, _ := message2.Pack()
		pack := append(pack1, pack2...)
		if _, err := conn.Write(pack); err != nil {
			fmt.Println("client write err:", err.Error())
			continue
		}
		fmt.Println("client write successfully")
		buf := make([]byte, 512)
		if _, err := conn.Read(buf); err != nil {
			fmt.Println("client read err", err.Error())
			continue
		}
		fmt.Println("client read successfully:", string(buf))
		time.Sleep(1 * time.Second)
	}
	runtime.Goexit()
}

func TestHook(t *testing.T) {
	conn, err := net.Dial("tcp", "0.0.0.0:8099")
	if err != nil {
		t.Error(err)
	}
	for {
		buf := make([]byte, 512)
		if _, err := conn.Read(buf); err != nil {
			fmt.Println("client read err", err.Error())
			continue
		}
		fmt.Println("client read successfully:", string(buf))
		time.Sleep(1 * time.Second)
	}
}

func TestBottleNeck(t *testing.T) {
	conn, err := net.Dial("tcp", "0.0.0.0:8099")
	if err != nil {
		t.Error(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i := 0; i < 100000; {
			buf := make([]byte, 512)
			if _, err := conn.Read(buf); err != nil {
				fmt.Println("client read err", err.Error())
				return
			}
			fmt.Println("client read successfully:", string(buf))
			i++
		}
		defer wg.Done()
	}()

	for i := 0; i < 100000; {
		message := mnet.NewMessage([]byte("Hello ginm"), 10)
		message.SetMsgID(2)
		pack, _ := message.Pack()
		if _, err := conn.Write(pack); err != nil {
			fmt.Println("client read err", err.Error())
			continue
		}
		i++
	}
	wg.Wait()
}

func TestHeartBeat(t *testing.T) {
	conn, err := net.Dial("tcp", "0.0.0.0:8099")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("client start successfully")
	for {
		buf := make([]byte, 512)
		if _, err := conn.Read(buf); err != nil {
			fmt.Println("client read err", err.Error())
			continue
		}
		fmt.Println("client read successfully:", string(buf))

		message1 := mnet.NewMessage([]byte("pong"), 4)
		message1.SetMsgID(inter.HeartBeatDefaultMsgID)
		pack, _ := message1.Pack()
		if _, err := conn.Write(pack); err != nil {
			fmt.Println("client write err:", err.Error())
			continue
		}
		fmt.Println("client write successfully")
	}
	runtime.Goexit()
}

func TestKcp(t *testing.T) {
	conn, err := kcp.Dial("0.0.0.0:8099")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("client start successfully")
	for {
		message1 := mnet.NewMessage([]byte("2Hello ginm"), 11)
		message1.SetMsgID(2)
		pack1, _ := message1.Pack()
		message2 := mnet.NewMessage([]byte("2Hello ginm2"), 12)
		message2.SetMsgID(2)
		pack2, _ := message2.Pack()
		pack := append(pack1, pack2...)
		if _, err := conn.Write(pack); err != nil {
			fmt.Println("client write err:", err.Error())
			continue
		}
		fmt.Println("client write successfully")
		buf := make([]byte, 512)
		if _, err := conn.Read(buf); err != nil {
			fmt.Println("client read err", err.Error())
			continue
		}
		fmt.Println("client read successfully:", string(buf))
		time.Sleep(1 * time.Second)
	}
	runtime.Goexit()
}

func TestSliceRouter(t *testing.T) {
	conn, err := net.Dial("tcp", "0.0.0.0:8099")
	if err != nil {
		fmt.Println("client start err, exit!", err)
		return
	}

	message1 := mnet.NewMessage([]byte("Hello ginm"), 10)
	message1.SetMsgID(1)
	pack1, _ := message1.Pack()
	_, err = conn.Write(pack1)
	if err != nil {
		fmt.Println("write error err ", err)
		return
	}
	fmt.Println("client write successfully")
	buf := make([]byte, 512)
	if _, err := conn.Read(buf); err != nil {
		fmt.Println("client read err", err.Error())
		return
	}
	fmt.Println("client read successfully:", string(buf))
	time.Sleep(1 * time.Second)
}

func TestAsynOp(t *testing.T) {
	conn, err := net.Dial("tcp", "0.0.0.0:8099")
	if err != nil {
		fmt.Println("client start err, exit!", err)
		return
	}

	message1 := mnet.NewMessage([]byte("Hello ginm"), 10)
	message1.SetMsgID(1)
	pack1, _ := message1.Pack()
	_, err = conn.Write(pack1)
	if err != nil {
		fmt.Println("write error err ", err)
		return
	}
	fmt.Println("client write successfully")
	buf := make([]byte, 512)
	if _, err := conn.Read(buf); err != nil {
		fmt.Println("client read err", err.Error())
		return
	}
	fmt.Println("client read successfully:", string(buf))
	time.Sleep(1 * time.Second)
}
func TestIntercept(t *testing.T) {
	conn, err := net.Dial("tcp", "0.0.0.0:8099")
	if err != nil {
		fmt.Println("client start err, exit!", err)
		return
	}

	message1 := mnet.NewMessage([]byte("Hello ginm"), 10)
	message1.SetMsgID(1)
	pack1, _ := message1.Pack()
	_, err = conn.Write(pack1)
	if err != nil {
		fmt.Println("write error err ", err)
		return
	}
	fmt.Println("client write successfully")
	buf := make([]byte, 512)
	if _, err := conn.Read(buf); err != nil {
		fmt.Println("client read err", err.Error())
		return
	}
	fmt.Println("client read successfully:", string(buf))
	time.Sleep(1 * time.Second)
}

func TestDecoder(t *testing.T) {
	conn, err := net.Dial("tcp", "0.0.0.0:8099")
	if err != nil {
		fmt.Println("client start err, exit!", err)
		return
	}

	message1 := mnet.NewMessage([]byte("Hello ginm"), 10)
	message1.SetMsgID(0x10)
	pack1, _ := message1.Pack()
	message2 := mnet.NewMessage([]byte("Hello ginm"), 10)
	message2.SetMsgID(0x10)
	pack2, _ := message2.Pack()
	pack1 = append(pack1, pack2...)
	_, err = conn.Write(pack1)
	if err != nil {
		fmt.Println("write error err ", err)
		return
	}
	fmt.Println("client write successfully")
	buf := make([]byte, 512)
	if _, err := conn.Read(buf); err != nil {
		fmt.Println("client read err", err.Error())
		return
	}
	fmt.Println("client read successfully:", string(buf))
	time.Sleep(1 * time.Second)
}

func TestDecoderWebsocket(t *testing.T) {
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial("ws://0.0.0.0:8099", nil)
	if err != nil {
		fmt.Println("client start err, exit!", err)
		return
	}

	message1 := mnet.NewMessage([]byte("Hello ginm1"), 11)
	message1.SetMsgID(0x10)
	pack1, _ := message1.Pack()
	err = conn.WriteMessage(1, pack1)
	if err != nil {
		fmt.Println("write error err ", err)
		return
	}
	fmt.Println("client write successfully")
	_, p, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("client read err", err.Error())
		return
	}
	fmt.Println("client read successfully:", string(p))
	time.Sleep(1 * time.Second)
}