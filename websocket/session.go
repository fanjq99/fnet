package websocket

import (
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/fanjq99/fnet"
	"github.com/fanjq99/glog"
	"github.com/gorilla/websocket"
)

type Session struct {
	Conn      *websocket.Conn
	closed    int32
	sessionId int64

	tag     interface{}
	tagLock sync.RWMutex

	sendChan chan []byte

	wg sync.WaitGroup

	OnClose func() // 关闭函数回调
}

func NewSession(conn *websocket.Conn) *Session {
	return &Session{
		Conn:      conn,
		closed:    -1,
		sessionId: 0,
		sendChan:  make(chan []byte, 100),
	}
}

func (s *Session) RemoteAddr() string {
	if s.Conn == nil {
		return ""
	}
	return s.Conn.RemoteAddr().String()
}

func (s *Session) GetID() int64 {
	return s.sessionId
}

func (s *Session) SetID(id int64) {
	s.sessionId = id
}

func (s *Session) SetTag(tag interface{}) {
	s.tagLock.Lock()
	defer s.tagLock.Unlock()

	s.tag = tag
}

func (s *Session) GetTag() interface{} {
	s.tagLock.RLock()
	s.tagLock.RUnlock()
	return s.tag
}

func (s *Session) Start() {
	if atomic.CompareAndSwapInt32(&s.closed, -1, 0) {
		s.wg.Add(2)
		go s.sendLoop()
		go s.receiveLoop()
		glog.Infoln("[连接] 收到连接 ", s.RemoteAddr())
	}
}

func (s *Session) Close() bool {
	if s.IsClosed() {
		return false
	}
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		glog.Infoln("[连接] 断开连接 ", s.GetID())
		s.OnClose()
		close(s.sendChan)
		s.Conn.Close()
		s.wg.Wait()
	}

	return true
}

func (s *Session) IsClosed() bool {
	return atomic.LoadInt32(&s.closed) != 0
}

func (s *Session) send(buffer []byte) bool {
	if s.IsClosed() {
		glog.Errorln("ws session closed", s.GetID())
		return false
	}

	s.sendChan <- buffer
	return true
}

func (s *Session) receiveLoop() {
	defer func() {
		if err := recover(); err != nil {
			glog.Errorln("[异常] ", err, "\n", string(debug.Stack()))
		}
		s.Close()
		s.wg.Done()
	}()

	for {
		// 读超时
		t, raw, err := s.Conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				glog.Errorln("receiveLoop error", err)
			}
			return
		}

		switch t {
		case websocket.TextMessage:
			// FIXME: 暂时不用
		case websocket.BinaryMessage:
			fnet.DecodeMessage(s, raw)
		case websocket.CloseMessage:
			glog.Infoln("receiveLoop closeMessage")
			return
		}
	}
}

func (s *Session) sendLoop() {
	defer func() {
		if err := recover(); err != nil {
			glog.Errorln("[异常] ", err, "\n", string(debug.Stack()))
		}
		s.Close()
		s.wg.Done()
	}()

	for {
		select {
		case data, ok := <-s.sendChan:
			if !ok {
				glog.Infoln("close web socket", s.GetID())
				s.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}
			s.Conn.WriteMessage(websocket.BinaryMessage, data)
		}
	}
}

func (s *Session) Reply(header *fnet.Message, msg interface{}, status uint8) bool {
	buf := fnet.ReplyMessage(header, msg, status)
	if buf == nil {
		glog.Errorln("Reply buf is nil", s.GetID())
		return false
	}

	return s.send(buf)
}

func (s *Session) Request(msg interface{}, status, serializeType uint8, indexId uint32) bool {
	buf := fnet.RequestMessage(msg, status, serializeType, indexId)
	if buf == nil {
		glog.Errorln("Request buf is nil", s.GetID())
	}

	return s.send(buf)
}
