package socket

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fanjq99/fnet"
	"github.com/fanjq99/glog"
)

type TcpServer struct {
	listener *net.TCPListener
	maxConn  int

	beforeExit func() //服务退出前回调函数
	reload     func() //重新加载函数，主要为配置文件的重新加载

	closeHandler   []fnet.OnCloseFunc
	connectHandler []fnet.OnConnectFunc

	sessionMgr *fnet.SessionManager

	stoped bool
}

const (
	defaultMaxConn = 10000
)

func NewDefaultTcpServer(addr string) *TcpServer {
	return NewTcpServer(addr, defaultMaxConn)
}

func NewTcpServer(addr string, maxcon int) *TcpServer {
	s := &TcpServer{
		maxConn:    maxcon,
		beforeExit: func() {},
		reload:     func() {},
		sessionMgr: fnet.NewSessionManager(),
	}
	err := s.bind(addr)
	if err != nil {
		return nil
	} else {
		return s
	}
}

func (t *TcpServer) SetReload(f func()) {
	t.reload = f
}

func (t *TcpServer) SetBeforeExit(f func()) {
	t.beforeExit = f
}

func (t *TcpServer) AddCloseHandler(f fnet.OnCloseFunc) {
	t.closeHandler = append(t.closeHandler, f)
}

func (t *TcpServer) AddConnectHandler(f fnet.OnConnectFunc) {
	t.connectHandler = append(t.connectHandler, f)
}

func (t *TcpServer) bind(address string) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		glog.Errorln(err)
		return err
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		glog.Errorln(err)
		return err
	}

	glog.Infoln("tcp server address:", address)
	t.listener = listener
	return nil
}

// Run run with signal handle
func (t *TcpServer) Run() {
	t.handleSignals()
	t.Start()
}

func (t *TcpServer) Start() {
	for {
		if t.stoped {
			break
		}
		t.listener.SetDeadline(time.Now().Add(time.Second * 1))

		conn, err := t.listener.AcceptTCP()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				//glog.Errorln("accept error timeout", err)
				continue
			}
			glog.Errorln("accept error", err)
			continue
		}

		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(1 * time.Minute)
		conn.SetNoDelay(true)
		conn.SetWriteBuffer(128 * 1024)
		conn.SetReadBuffer(128 * 1024)

		//max client exceed
		if t.sessionMgr.Len() >= t.maxConn {
			glog.Errorln("tcp conn reach max num:", t.maxConn)
			conn.Close()
		} else {
			go t.handleConnection(conn)
		}
	}
}

func (t *TcpServer) handleConnection(conn *net.TCPConn) {
	ses := NewSession(conn)
	ses.OnClose = func() {
		t.sessionMgr.Remove(ses)
		for _, f := range t.closeHandler {
			f(ses)
		}
	}

	t.sessionMgr.Add(ses)
	for _, f := range t.connectHandler {
		f(ses)
	}

	ses.Start()
}

func (t *TcpServer) handleSignals() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for sig := range ch {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				glog.Infof("receive stop signal: [%s]", sig)
				t.Stop()
			case syscall.SIGHUP:
				glog.Infof("receive signal: [%s]", sig)
				// 热加载预留
				t.reload()
			default:
				glog.Infof("Received other signal [%s]", sig)
			}
		}
	}()
}

func (t *TcpServer) Stop() {
	if t.stoped {
		return
	}

	glog.Infoln("server stop start")
	t.stoped = true
	t.listener.Close()

	t.sessionMgr.CloseAll()
	t.beforeExit()
	glog.Infoln("server stop end")
}
