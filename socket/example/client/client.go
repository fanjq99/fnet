package main

import (
	"flag"
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/fanjq99/fnet/socket"
	"github.com/fanjq99/fnet/socket/example/protocol"
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

	address := "127.0.0.1:12002"
	client := socket.NewTcpClient("client", address)
	if client == nil {
		fmt.Println("client connect error")
	}

	client.Run()

	data := &protocol.TestData{
		Name: proto.String("test1"),
		Id:   proto.Uint32(123456),
	}
	test := &protocol.C2S_TestRequest{
		Data: data,
	}

	for i := 1; i < 1000; i++ {
		go func() {
			client.Send(test, 0)
		}()
	}

	ch := make(chan struct{})
	<-ch
}
