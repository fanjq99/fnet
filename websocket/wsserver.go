package websocket

import (
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/fanjq99/fnet"
	"github.com/fanjq99/glog"
	"github.com/gorilla/websocket"
)

type WSServer struct {
	MaxConn    int
	sessionMgr *fnet.SessionManager

	closeHandler   []fnet.OnCloseFunc
	connectHandler []fnet.OnConnectFunc

	address string

	beforeExit func()
	reload     func()

	wg sync.WaitGroup
}

const (
	defaultMaxConn = 10000
)

// use default options
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewDefaultServer(addr string) *WSServer {
	return NewServer(addr, defaultMaxConn)
}

func NewServer(addr string, maxcon int) *WSServer {
	s := &WSServer{
		MaxConn:    maxcon,
		sessionMgr: fnet.NewSessionManager(),
		address:    addr,
		beforeExit: func() {},
		reload:     func() {},
	}
	return s
}

func (ws *WSServer) SetReload(f func()) {
	ws.reload = f
}

func (ws *WSServer) SetBeforeExit(f func()) {
	ws.beforeExit = f
}

func (ws *WSServer) AddCloseHandler(f fnet.OnCloseFunc) {
	ws.closeHandler = append(ws.closeHandler, f)
}

func (ws *WSServer) AddConnectHandler(f fnet.OnConnectFunc) {
	ws.connectHandler = append(ws.connectHandler, f)
}

func (ws *WSServer) Start() {
	u, err := url.Parse(ws.address)
	glog.Infof("websocket server url path%s, host%s\n", u.Path, u.Host)

	if err != nil {
		glog.Errorln("WSServer start error", err, ws.address)
		return
	}

	if u.Path == "" {
		glog.Errorln("WSServer: expect path in url to listen", ws.address)
		return
	}

	http.HandleFunc(u.Path, func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		go ws.handleConnection(c)
	})

	err = http.ListenAndServe(u.Host, nil)
	if err != nil {
		glog.Errorln(err)
	}
}

func (ws *WSServer) handleConnection(conn *websocket.Conn) {
	ses := NewSession(conn)
	ses.OnClose = func() {
		ws.sessionMgr.Remove(ses)
		for _, f := range ws.closeHandler {
			f(ses)
		}

	}
	ws.sessionMgr.Add(ses)
	for _, f := range ws.connectHandler {
		f(ses)
	}
	ses.Start()
}

func (ws *WSServer) Run() {
	ws.handleSignals()
	ws.Start()
}

func (ws *WSServer) handleSignals() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for sig := range ch {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				glog.Infof("receive stop signal: [%s]", sig)
				ws.stop()
			case syscall.SIGHUP:
				// 加载预留
				glog.Infof("receive usr1 signal: [%s]", sig)
				ws.reload()
			default:
				glog.Infof("Received other signal [%s]", sig)
			}
		}
	}()
}

func (ws *WSServer) stop() {
	glog.Infoln("server stop start")
	ws.sessionMgr.CloseAll()
	ws.beforeExit()
	glog.Infoln("server stop end")

	os.Exit(0)
}
