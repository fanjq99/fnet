package main

import (
	"flag"
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/fanjq99/fnet"
	"github.com/fanjq99/fnet/socket"
	_ "github.com/fanjq99/fnet/socket/example/protocol"
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

	address := ":12002"
	server := socket.NewTcpServer(address, 10)
	if server == nil {
		fmt.Println("server bind error")
		return
	}

	fnet.RegisterMessageFunction(handlerMessage)
	glog.Info("server start")
	fmt.Println("server start")
	server.Run()
}

func handlerMessage(session fnet.ISession, msg *fnet.Message) {
	glog.Infoln("receive msg", msg)
	//switch msg.MsgId {
	//case 1:
	//	//session.Reply(msg, )
	//}
}
