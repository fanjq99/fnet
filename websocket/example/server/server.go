package main

import (
	"flag"
	"fmt"
	"github.com/fanjq99/fnet"
	"runtime"
	"runtime/debug"

	"github.com/fanjq99/fnet/websocket"
	"github.com/fanjq99/glog"
)

func main() {
	flag.Parse()
	defer glog.Flush()
	defer func() {
		if err := recover(); err != nil {
			glog.Errorln("[异常] ", err, "\n", string(debug.Stack()))
		}
	}()

	runtime.GOMAXPROCS(runtime.NumCPU())

	address := "ws://127.0.0.1:8080/websocket"
	server := websocket.NewDefaultServer(address)
	if server == nil {
		fmt.Println("server create error")
	}

	server.AddConnectHandler(func(session fnet.ISession) {
		fmt.Println("ses connect", session)
	})

	server.AddCloseHandler(func(session fnet.ISession) {
		fmt.Println("ses onclose", session)
	})

	fnet.RegisterMessageFunction(handlerMessage)

	glog.Info("server start..........")
	server.Run()
}

func handlerMessage(session fnet.ISession, msg *fnet.Message) {
	fmt.Println("receive msg", session.GetID(), msg)
}
