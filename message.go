package fnet

import (
	"bytes"
	"encoding/binary"

	"github.com/fanjq99/fnet/codec"
	"github.com/fanjq99/glog"
)

var (
	// Codecs are codecs supported by rpc. You can add customized codecs in Codecs.
	Codecs = map[SerializeType]codec.Codec{
		ProtoBuffer: &codec.PBCodec{},
		JSON:        &codec.JSONCodec{},
	}
)

// SerializeType defines serialization type of Data.
type SerializeType byte

const (
	// ProtoBuffer for payload.
	ProtoBuffer SerializeType = iota
	// JSON for payload.
	JSON
)

// MessageType is message type of requests and resposnes.
type MessageType byte

const (
	// Request is message type of request
	Request MessageType = iota
	// Response is message type of response
	Response
)

// MessageStatusType is status of messages.
type MessageStatusType byte

const (
	// Normal is normal requests and responses.
	Normal MessageStatusType = iota
	// Error indicates some errors occur.
	Error
)

<<<<<<< HEAD
// Message tcp & webscoket share message struct
=======
>>>>>>> parent of 8c3b0af... full commit
type Message struct {
	MsgId         uint32
	Index         uint32
	SerializeType uint8
	MessageType   uint8
	Status        uint8
	GateSessionId int64
	Data          []byte
}

func (m *Message) Encode() []byte {
	buff := new(bytes.Buffer)
	binary.Write(buff, binary.LittleEndian, m.MsgId)
	binary.Write(buff, binary.LittleEndian, m.Index)
	binary.Write(buff, binary.LittleEndian, m.SerializeType)
	binary.Write(buff, binary.LittleEndian, m.MessageType)
	binary.Write(buff, binary.LittleEndian, m.Status)
	binary.Write(buff, binary.LittleEndian, m.GateSessionId)
	binary.Write(buff, binary.LittleEndian, m.Data)

	return buff.Bytes()
}

func (m *Message) Decode(buff []byte) error {
	buf := bytes.NewBuffer(buff)
	err := binary.Read(buf, binary.LittleEndian, &m.MsgId)
	if err != nil {
		glog.Errorln("decode MsgId err:", err)
		return err
	}
	err = binary.Read(buf, binary.LittleEndian, &m.Index)
	if err != nil {
		glog.Errorln("decode Index err:", err)
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &m.SerializeType)
	if err != nil {
		glog.Errorln("decode SerializeType err:", err)
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &m.MessageType)
	if err != nil {
		glog.Errorln("decode MessageType err:", err)
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &m.Status)
	if err != nil {
		glog.Errorln("decode Status err:", err)
		return err
	}

	err = binary.Read(buf, binary.LittleEndian, &m.GateSessionId)
	if err != nil {
		glog.Errorln("decode GateSessionId err:", err)
		return err
	}

	m.Data = buff[19:]

	return nil
}
