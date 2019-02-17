package socket

import (
	"io"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fanjq99/fnet"
	"github.com/fanjq99/glog"
)

const (
	cmdMaxSize    = 1024 * 1024
	cmdHeaderSize = 4
	cmdVerifytime = 15
)

type Session struct {
	Conn      *net.TCPConn
	closed    int32
	sessionId int64
	recvBuff  *ByteBuffer
	sendChan  chan []byte
	HeartTime time.Time

	tag     interface{}
	taglock sync.RWMutex

	waitGroup sync.WaitGroup
	OnClose   func() // 关闭函数回调
}

func NewSession(conn *net.TCPConn) *Session {
	return &Session{
		Conn:      conn,
		closed:    -1,
		sessionId: 0,
		recvBuff:  NewByteBuffer(),
		sendChan:  make(chan []byte, 10),
		HeartTime: time.Now().Add(cmdVerifytime * time.Second),
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
	s.taglock.Lock()
	defer s.taglock.Unlock()

	s.tag = tag
}

func (s *Session) GetTag() interface{} {
	s.taglock.RLock()
	s.taglock.RUnlock()
	return s.tag
}

func (s *Session) Start() {
	if atomic.CompareAndSwapInt32(&s.closed, -1, 0) {
		s.waitGroup.Add(2)
		go s.sendloop()
		go s.recvloop()
		glog.Infoln("[tcp连接] 收到连接 ", s.RemoteAddr())
	}
}

func (s *Session) Close() bool {
	if s.IsClosed() {
		return false
	}
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		glog.Infoln("[tcp连接] 断开连接 ", s.RemoteAddr())
		s.Conn.Close()
		close(s.sendChan)
		s.recvBuff.Reset()
		s.OnClose()
		s.waitGroup.Wait()
	}

	return true
}

func (s *Session) IsClosed() bool {
	return atomic.LoadInt32(&s.closed) != 0
}

func (s *Session) recvloop() {
	defer func() {
		if err := recover(); err != nil {
			glog.Errorln("[异常] ", err, "\n", string(debug.Stack()))
		}
		s.Close()
		s.waitGroup.Done()
	}()

	var (
		neednum   int
		readnum   int
		err       error
		totalsize int
		datasize  int
		msgbuff   []byte
	)

	for {
		totalsize = s.recvBuff.RdSize()

		if totalsize < cmdHeaderSize {
			neednum = cmdHeaderSize - totalsize
			if s.recvBuff.WrSize() < neednum {
				s.recvBuff.WrGrow(neednum)
			}

			readnum, err = io.ReadAtLeast(s.Conn, s.recvBuff.WrBuf(), neednum)
			if err != nil {
				glog.Errorf("[tcp recvloop] ip(%s) read data error(%v) ", s.RemoteAddr(), err)
				return
			}

			s.recvBuff.WrFlip(readnum)
			totalsize = s.recvBuff.RdSize()
		}

		msgbuff = s.recvBuff.RdBuf()

		datasize = int(msgbuff[0]) | int(msgbuff[1])<<8 | int(msgbuff[2])<<16 | int(msgbuff[3])<<24
		if datasize > cmdMaxSize {
			glog.Errorf("[tcp recvloop] ip(%s) datasize(%d) > cmdMaxSize(128*1024) ", s.RemoteAddr(), datasize)
			return
		}

		if totalsize < cmdHeaderSize+datasize {

			neednum = cmdHeaderSize + datasize - totalsize
			if s.recvBuff.WrSize() < neednum {
				s.recvBuff.WrGrow(neednum)
			}

			readnum, err = io.ReadAtLeast(s.Conn, s.recvBuff.WrBuf(), neednum)
			if err != nil {
				glog.Errorf("[tcp recvloop] ip(%s) read data error(%v) ", s.RemoteAddr(), err)
				return
			}

			s.recvBuff.WrFlip(readnum)
			msgbuff = s.recvBuff.RdBuf()
		}

		fnet.DecodeMessage(s, msgbuff[cmdHeaderSize:cmdHeaderSize+datasize])
		s.recvBuff.RdFlip(cmdHeaderSize + datasize)
	}
}

func (s *Session) sendloop() {
	var (
		tmpByte  = NewByteBuffer()
		timeout  = time.NewTimer(cmdVerifytime * time.Second)
		writenum int
		err      error
	)

	defer func() {
		if err := recover(); err != nil {
			glog.Errorln("[异常] ", err, "\n", string(debug.Stack()))
		}
		s.Close()
		timeout.Stop()
		s.waitGroup.Done()
	}()

	for {
		select {
		case buf, ok := <-s.sendChan:
			if !ok {
				glog.Infoln("tcp session close sendChan", s.GetID())
				return
			}
			tmpByte.Append(buf...)
			for {
				if !tmpByte.RdReady() {
					break
				}
				writenum, err = s.Conn.Write(tmpByte.RdBuf()[:tmpByte.RdSize()])
				if err != nil {
					glog.Errorf("[tcp sendloop] ip(%s) send error(%v)", s.RemoteAddr(), err)
					return
				}
				tmpByte.RdFlip(writenum)
			}
		case <-timeout.C:
			if s.HeartTime.Unix() < time.Now().Unix() {
				glog.Infoln("tcp timeout", s.RemoteAddr())
				return
			}
		}
	}
}

func (s *Session) send(buffer []byte) bool {
	if s.IsClosed() {
		glog.Errorln("session closed", s.GetID())
		return false
	}

	bsize := len(buffer)
	buf := make([]byte, 0, 4+bsize)
	header := []byte{byte(bsize), byte(bsize >> 8), byte(bsize >> 16), byte(bsize >> 24)}
	buf = append(buf, header...)
	buf = append(buf, buffer...)
	s.sendChan <- buf

	return true
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
