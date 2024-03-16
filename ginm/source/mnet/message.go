package mnet

import (
	"bytes"
	"encoding/binary"
	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
	"io"
	"mmo/ginm/pkg/utils"
	"mmo/ginm/source/decoder"
	"mmo/ginm/source/inter"
	"net"
)

type message struct {
	data    []byte
	dataLen uint32
	msgType uint32
}

func (m *message) SetData(value []byte) {
	m.data = value
}

func (m *message) SetDataLen(length uint32) {
	m.dataLen = length
}

func (m *message) SetMsgID(msgID uint32) {
	m.msgType = msgID
}

func NewMessage(data []byte, dataLen uint32) inter.Message {
	return &message{data: data, dataLen: dataLen}
}

func (m *message) PackHtlvCrc() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	if err := buf.WriteByte(0xA1); err != nil {
		return nil, utils.Wrap(err, " ")
	}
	if err := binary.Write(buf, binary.BigEndian, uint8(m.msgType)); err != nil {
		return nil, utils.Wrap(err, " ")
	}
	if err := binary.Write(buf, binary.BigEndian, uint8(m.dataLen)); err != nil {
		return nil, utils.Wrap(err, " ")
	}
	if err := binary.Write(buf, binary.BigEndian, m.data); err != nil {
		return nil, utils.Wrap(err, " ")
	}
	if err := binary.Write(buf, binary.BigEndian, decoder.GetCrc(buf.Bytes())); err != nil {
		return nil, utils.Wrap(err, " ")
	}
	return buf.Bytes(), nil
}

func (m *message) Pack() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	if err := binary.Write(buf, binary.BigEndian, m.msgType); err != nil {
		return nil, utils.Wrap(err, " ")
	}
	if err := binary.Write(buf, binary.BigEndian, m.dataLen); err != nil {
		return nil, utils.Wrap(err, " ")
	}
	if err := binary.Write(buf, binary.BigEndian, m.data); err != nil {
		return nil, utils.Wrap(err, " ")
	}
	return buf.Bytes(), nil
}

func (m *message) Unpack(tcpConn net.Conn, wsConn *websocket.Conn, kcpConn *kcp.UDPSession) error {
	if tcpConn != nil {
		msgType := make([]byte, 4)
		if _, err := io.ReadFull(tcpConn, msgType); err != nil {
			return utils.Wrap(err, "readMsgType failed")
		}
		dataLen := make([]byte, 4)
		if _, err := io.ReadFull(tcpConn, dataLen); err != nil {
			return utils.Wrap(err, "readDataLen failed")
		}
		m.dataLen = binary.BigEndian.Uint32(dataLen)
		m.msgType = binary.BigEndian.Uint32(msgType)
		m.data = make([]byte, m.dataLen)
		if _, err := io.ReadFull(tcpConn, m.data); err != nil {
			return utils.Wrap(err, "readData failed")
		}
	}
	//if wsConn != nil {
	//	dataLen := make([]byte, 4)
	//	if _, err := io.ReadFull(tcpConn, dataLen); err != nil {
	//		return utils.Wrap(err, "readDataLen failed")
	//	}
	//	m.dataLen = binary.LittleEndian.Uint32(dataLen)
	//	msgType := make([]byte, 4)
	//	if _, err := io.ReadFull(tcpConn, msgType); err != nil {
	//		return utils.Wrap(err, "readMsgType failed")
	//	}
	//	m.msgType = binary.LittleEndian.Uint32(msgType)
	//	m.data = make([]byte, m.dataLen)
	//	if _, err := io.ReadFull(tcpConn, m.data); err != nil {
	//		return utils.Wrap(err, "readData failed")
	//	}
	//}
	if kcpConn != nil {
		msgType := make([]byte, 4)
		if _, err := io.ReadFull(tcpConn, msgType); err != nil {
			return utils.Wrap(err, "readMsgType failed")
		}
		dataLen := make([]byte, 4)
		if _, err := io.ReadFull(tcpConn, dataLen); err != nil {
			return utils.Wrap(err, "readDataLen failed")
		}
		m.dataLen = binary.BigEndian.Uint32(dataLen)
		m.msgType = binary.BigEndian.Uint32(msgType)
		m.data = make([]byte, m.dataLen)
		if _, err := io.ReadFull(tcpConn, m.data); err != nil {
			return utils.Wrap(err, "readData failed")
		}
	}
	return nil
}

func (m *message) GetData() []byte {
	return m.data
}

func (m *message) GetDataLen() uint32 {
	return m.dataLen
}

func (m *message) GetHeaderLen() uint32 {
	return 8
}

func (m *message) GetMsgType() uint32 {
	return m.msgType
}
