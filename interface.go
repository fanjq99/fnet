package fnet

type ISession interface {
	GetID() int64
	SetID(int64)
	Close() bool
	Reply(header *Message, msg interface{}, status uint8) bool
	Request(msg interface{}, status, serializeType uint8, indexId uint32) bool
}

type MessageFunc func(session ISession, msg *Message)

type OnCloseFunc func(session ISession)
type OnConnectFunc func(session ISession)
