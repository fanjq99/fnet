package websocket

import (
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fanjq99/fnet"
	"github.com/fanjq99/glog"
	"github.com/gorilla/websocket"
)

var (
	reconnectInterval = 200 //ms
	maxReconnectSec   = 60 * 1000
)

type WSClient struct {
	sessionMgr *fnet.SessionManager

	running      bool // 运行状态
	runningGuard sync.RWMutex

	Address string
	Name    string
	ses     *Session
	stop    bool

	closeHandler   []fnet.OnCloseFunc
	connectHandler []fnet.OnConnectFunc

	SerializeType fnet.SerializeType

	indexID uint32
}

func NewWSClient(name, address string) *WSClient {
	return &WSClient{
		sessionMgr:    fnet.NewSessionManager(),
		running:       false,
		Address:       address,
		Name:          name,
		stop:          false,
		SerializeType: fnet.ProtoBuffer,
	}
}

func (w *WSClient) IsRunning() bool {
	w.runningGuard.RLock()
	defer w.runningGuard.RUnlock()

	return w.running
}

func (w *WSClient) SetRunning(v bool) {
	w.runningGuard.Lock()
	defer w.runningGuard.Unlock()

	w.running = v
}

func (w *WSClient) GetID() int64 {
	if w.ses != nil {
		return w.ses.GetID()
	}
	return 0
}

func (w *WSClient) AddCloseHandler(f fnet.OnCloseFunc) {
	w.closeHandler = append(w.closeHandler, f)
}

func (w *WSClient) AddConnectHandler(f fnet.OnConnectFunc) {
	w.connectHandler = append(w.connectHandler, f)
}

func (w *WSClient) Run() {
	if w.IsRunning() {
		return
	}

	rawURL, err := url.Parse(w.Address)

	if err != nil {
		glog.Errorln(err, w.Address)
		return
	}

	if rawURL.Path == "" {
		glog.Errorln("webSocket: expect path in url to listen", w.Address)
		return
	}

	err = w.connect(w.Address)
	if err != nil {
		panic(err)
	}
}

func (w *WSClient) reconnect() {
	if !w.IsRunning() {
		return
	}

	var reconnectTime = 1
	for {
		err := w.connect(w.Address)
		if err != nil {
			glog.Errorln("reconnect failed:", err)
		} else {
			break
		}

		interval := reconnectTime * reconnectInterval
		if interval > maxReconnectSec {
			interval = maxReconnectSec
		}

		time.Sleep(time.Duration(interval) * time.Millisecond)
		maxReconnectSec++
	}
}

func (w *WSClient) connect(address string) error {
	conn, _, err := websocket.DefaultDialer.Dial(w.Address, nil)
	if err != nil {
		return err
	}

	ses := NewSession(conn)
	ses.OnClose = func() {
		w.sessionMgr.Remove(ses)
		for _, f := range w.closeHandler {
			f(ses)
		}
		w.reconnect()
	}
	w.sessionMgr.Add(ses)
	ses.Start()
	w.ses = ses
	w.SetRunning(true)
	for _, f := range w.connectHandler {
		f(ses)
	}

	return nil
}

func (w *WSClient) Stop() {
	if !w.IsRunning() {
		return
	}

	w.SetRunning(false)

	if w.ses != nil {
		w.ses.Close()
	}

	glog.Infoln("ws client close ", w.Name, w.Address)
}

func (w *WSClient) Reply(header *fnet.Message, msg interface{}, status uint8) bool {
	if !w.IsRunning() {
		return false
	}

	buf := fnet.ReplyMessage(header, msg, status)
	if buf == nil {
		return false
	}

	for {
		if w.ses.send(buf) {
			break
		}
		time.Sleep(time.Duration(reconnectInterval) * time.Millisecond)
	}

	return true
}

func (w *WSClient) Send(msg interface{}, status uint8) bool {
	if !w.IsRunning() {
		return false
	}

	buf := fnet.RequestMessage(msg, status, uint8(w.SerializeType), atomic.AddUint32(&w.indexID, 1))
	if buf == nil {
		return false
	}
	fmt.Println("send data:", buf)

	for {
		if w.ses.send(buf) {
			break
		}
		time.Sleep(time.Duration(reconnectInterval) * time.Millisecond)
	}

	return true
}
