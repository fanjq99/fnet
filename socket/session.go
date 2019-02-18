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
	Conn        *net.TCPConn
	closed      int32
	sessionId   int64
	receiveBuff *ByteBuffer
	sendChan    chan []byte
	HeartTime   time.Time

	tag     interface{}
	tagLock sync.RWMutex

	waitGroup sync.WaitGroup
	OnClose   func() // 关闭函数回调
}

func NewSession(conn *net.TCPConn) *Session {
	return &Session{
		Conn:        conn,
		closed:      -1,
		sessionId:   0,
		receiveBuff: NewByteBuffer(),
		sendChan:    make(chan []byte, 10),
		HeartTime:   time.Now().Add(cmdVerifytime * time.Second),
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
		s.waitGroup.Add(2)
		go s.sendLoop()
		go s.receiveLoop()
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
		s.receiveBuff.Reset()
		s.OnClose()
		s.waitGroup.Wait()
	}

	return true
}

func (s *Session) IsClosed() bool {
	return atomic.LoadInt32(&s.closed) != 0
}

func (s *Session) receiveLoop() {
	defer func() {
		if err := recover(); err != nil {
			glog.Errorln("[异常] ", err, "\n", string(debug.Stack()))
		}
		s.Close()
		s.waitGroup.Done()
	}()

	var (
		needNum   int
		readNum   int
		err       error
		totalSize int
		dataSize  int
		msgBuff   []byte
	)

	for {
		totalSize = s.receiveBuff.RdSize()

		if totalSize < cmdHeaderSize {
			needNum = cmdHeaderSize - totalSize
			if s.receiveBuff.WrSize() < needNum {
				s.receiveBuff.WrGrow(needNum)
			}

			readNum, err = io.ReadAtLeast(s.Conn, s.receiveBuff.WrBuf(), needNum)
			if err != nil {
				glog.Errorf("[tcp receiveLoop] ip(%s) read data error(%v) ", s.RemoteAddr(), err)
				return
			}

			s.receiveBuff.WrFlip(readNum)
			totalSize = s.receiveBuff.RdSize()
		}

		msgBuff = s.receiveBuff.RdBuf()

		dataSize = int(msgBuff[0]) | int(msgBuff[1])<<8 | int(msgBuff[2])<<16 | int(msgBuff[3])<<24
		if dataSize > cmdMaxSize {
			glog.Errorf("[tcp receiveLoop] ip(%s) dataSize(%d) > cmdMaxSize(128*1024) ", s.RemoteAddr(), dataSize)
			return
		}

		if totalSize < cmdHeaderSize+dataSize {

			needNum = cmdHeaderSize + dataSize - totalSize
			if s.receiveBuff.WrSize() < needNum {
				s.receiveBuff.WrGrow(needNum)
			}

			readNum, err = io.ReadAtLeast(s.Conn, s.receiveBuff.WrBuf(), needNum)
			if err != nil {
				glog.Errorf("[tcp receiveLoop] ip(%s) read data error(%v) ", s.RemoteAddr(), err)
				return
			}

			s.receiveBuff.WrFlip(readNum)
			msgBuff = s.receiveBuff.RdBuf()
		}

		fnet.DecodeMessage(s, msgBuff[cmdHeaderSize:cmdHeaderSize+dataSize])
		s.receiveBuff.RdFlip(cmdHeaderSize + dataSize)
	}
}

func (s *Session) sendLoop() {
	var (
		tmpByte  = NewByteBuffer()
		timeout  = time.NewTimer(cmdVerifytime * time.Second)
		writeNum int
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
				writeNum, err = s.Conn.Write(tmpByte.RdBuf()[:tmpByte.RdSize()])
				if err != nil {
					glog.Errorf("[tcp sendLoop] ip(%s) send error(%v)", s.RemoteAddr(), err)
					return
				}
				tmpByte.RdFlip(writeNum)
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

	bSize := len(buffer)
	buf := make([]byte, 0, 4+bSize)
	header := []byte{byte(bSize), byte(bSize >> 8), byte(bSize >> 16), byte(bSize >> 24)}
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
