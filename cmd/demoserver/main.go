package main

import (
	"fmt"
	"github.com/xfwduke/raft/demoproto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"time"
)

type DemoServer struct {
	hbch chan struct{}
}

func (ds *DemoServer) KeepHeartBeat(context.Context, *demoproto.HeartBeatRequest) (*demoproto.HeartBeatResponse, error) {
	fmt.Printf("%s: heatbeat request received\n", time.Now())
	ds.hbch <- struct{}{}
	fmt.Printf("%s: timeout reset signal sent\n", time.Now())
	return &demoproto.HeartBeatResponse{}, nil
}

func (ds *DemoServer) StartTimeoutLoop() {
	ds.hbch = make(chan struct{})
	for {
		select {
		case <-ds.hbch:
			fmt.Printf("%s: heartbeat reset\n", time.Now())
		case <-time.After(2 * time.Second):
			fmt.Printf("%s: heartbeat timeout triggered\n", time.Now())
		}
	}
}

var lis net.Listener

func init() {
	var err error
	lis, err = net.Listen("tcp", "127.0.0.1:22222")
	if err != nil {
		panic(err)
	}
}

func main() {
	demoServer := DemoServer{}
	grpcServer := grpc.NewServer()
	demoproto.RegisterHeartBeatServer(grpcServer, &demoServer)

	go grpcServer.Serve(lis)
	demoServer.StartTimeoutLoop()
}
