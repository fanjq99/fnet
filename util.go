package fnet

import (
	"os"
	"reflect"

	"github.com/fanjq99/glog"
)

var logicFunc MessageFunc

// RegisterMessageFunction call outside
func RegisterMessageFunction(f MessageFunc) {
	logicFunc = f
}

func DecodeMessage(s ISession, data []byte) {
	message := new(Message)
	err := message.Decode(data)
	if err != nil {
		glog.Errorln("message decode error:", err)
	} else {
		if logicFunc == nil {
			panic("not register message function")
			os.Exit(1)
		}

		logicFunc(s, message)
	}
}

func ReplyMessage(header *Message, msg interface{}, status uint8) []byte {
	serializeType := SerializeType(header.SerializeType)
	if serializeType != ProtoBuffer && serializeType != JSON {
		glog.Errorln("serializeType not support", header.SerializeType)
		return nil
	}

	data, err := Codecs[serializeType].Encode(msg)
	if err != nil {
		glog.Errorln(err)
		return nil
	}

	rType := reflect.TypeOf(msg).Elem()
	meta := MessageMetaByType(rType)
	if meta == nil {
		glog.Errorln("can not find message", rType)
		return nil
	}

	message := new(Message)
	message.MsgId = meta.ID
	message.Index = header.Index
	message.SerializeType = header.SerializeType
	message.MessageType = uint8(Response)
	message.Status = status
	message.Data = data

	return message.Encode()
}

func RequestMessage(msg interface{}, status, serializeType uint8, indexID uint32) []byte {
	data, err := Codecs[SerializeType(serializeType)].Encode(msg)
	if err != nil {
		glog.Errorln(err)
		return nil
	}

	rType := reflect.TypeOf(msg).Elem()
	meta := MessageMetaByType(rType)
	if meta == nil {
		glog.Errorln("can not find message", rType)
		return nil
	}

	message := new(Message)
	message.MsgId = meta.ID
	message.Index = indexID
	message.SerializeType = serializeType
	message.MessageType = uint8(Request)
	message.Status = status
	message.Data = data

	return message.Encode()
}
