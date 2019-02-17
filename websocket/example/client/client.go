package main

import (
	"flag"
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/fanjq99/fnet"
	"github.com/fanjq99/fnet/websocket"
	"github.com/fanjq99/fnet/websocket/example/protocol"
	"github.com/fanjq99/glog"
	"github.com/golang/protobuf/proto"
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
	client := websocket.NewWSClient("client", address)
	if client == nil {
		fmt.Println("client connect error")
	}

	client.AddConnectHandler(func(session fnet.ISession) {
		fmt.Println("on connect")
	})

	client.AddCloseHandler(func(session fnet.ISession) {
		fmt.Println("on close")
	})

	client.Run()

	data := &protocol.TestData{
		Name: proto.String("test1"),
		Id:   proto.Uint32(123456),
	}
	test := &protocol.C2S_TestRequest{
		Data: data,
	}
	client.Send(test, 0)

	ch := make(chan struct{})
	<-ch
}
