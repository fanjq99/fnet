package socket

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fanjq99/fnet"
	"github.com/fanjq99/glog"
)

var (
	reconnectInterval = 200 //ms
	maxReconnectSec   = 60 * 1000
)

type TcpClient struct {
	running      bool // 运行状态
	runningGuard sync.RWMutex

	Address       string
	Tag           string
	SerializeType fnet.SerializeType
	indexID       uint32

	ses *Session

	closeHandler   []fnet.OnCloseFunc
	connectHandler []fnet.OnConnectFunc

	sessionMgr *fnet.SessionManager
}

func NewTcpClient(tag, address string) *TcpClient {
	return &TcpClient{
		running:       false,
		Address:       address,
		Tag:           tag,
		SerializeType: fnet.ProtoBuffer,
		sessionMgr:    fnet.NewSessionManager(),
	}
}

func (t *TcpClient) IsRunning() bool {
	t.runningGuard.RLock()
	defer t.runningGuard.RUnlock()

	return t.running
}

func (t *TcpClient) SetRunning(v bool) {
	t.runningGuard.Lock()
	defer t.runningGuard.Unlock()

	t.running = v
}

func (t *TcpClient) GetID() int64 {
	if t.ses != nil {
		return t.ses.GetID()
	}
	return 0
}

func (t *TcpClient) AddCloseHandler(f fnet.OnCloseFunc) {
	t.closeHandler = append(t.closeHandler, f)
}

func (t *TcpClient) AddConnectHandler(f fnet.OnConnectFunc) {
	t.connectHandler = append(t.connectHandler, f)
}

func (t *TcpClient) Run() {
	if t.IsRunning() {
		return
	}

	err := t.connect(t.Address)
	if err != nil {
		panic(err)
	}
}

func (t *TcpClient) reconnect() {
	if !t.IsRunning() {
		return
	}

	var reconnectTime = 1
	for {
		err := t.connect(t.Address)
		if err != nil {
			glog.Errorln("reconect failed:", err)
		} else {
			break
		}

		interval := reconnectTime * reconnectInterval
		if interval > maxReconnectSec {
			interval = maxReconnectSec
		}

		time.Sleep(time.Duration(interval) * time.Millisecond)
		reconnectTime++
	}
}

func (t *TcpClient) connect(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	con := conn.(*net.TCPConn)
	con.SetKeepAlive(true)
	con.SetKeepAlivePeriod(1 * time.Minute)
	con.SetNoDelay(true)
	con.SetWriteBuffer(128 * 1024)
	con.SetReadBuffer(128 * 1024)

	ses := NewSession(con)
	ses.OnClose = func() {
		t.sessionMgr.Remove(ses)
		for _, f := range t.closeHandler {
			f(ses)
		}
		t.reconnect()
	}

	t.sessionMgr.Add(ses)
	ses.Start()
	t.ses = ses
	t.SetRunning(true)

	for _, f := range t.connectHandler {
		f(ses)
	}

	return nil
}

func (t *TcpClient) Stop() {
	if !t.IsRunning() {
		return
	}

	t.SetRunning(false)

	if t.ses != nil {
		t.ses.Close()
		t.ses = nil
	}
	glog.Infoln("tcp client close: ", t.Tag, t.Address)
}

func (t *TcpClient) Reply(header *fnet.Message, msg interface{}, status uint8) bool {
	if !t.IsRunning() {
		return false
	}

	buf := fnet.ReplyMessage(header, msg, status)
	if buf == nil {
		return false
	}

	for {
		if t.ses.send(buf) {
			break
		}
		time.Sleep(time.Duration(reconnectInterval) * time.Millisecond)
	}

	return true
}

func (t *TcpClient) Send(msg interface{}, status uint8) bool {
	if !t.IsRunning() {
		return false
	}

	buf := fnet.RequestMessage(msg, status, uint8(t.SerializeType), atomic.AddUint32(&t.indexID, 1))
	if buf == nil {
		return false
	}

	for {
		if t.ses.send(buf) {
			break
		}
		time.Sleep(time.Duration(reconnectInterval) * time.Millisecond)
	}

	return true
}
