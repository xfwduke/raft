package main

import (
	"context"
	"fmt"
	"github.com/xfwduke/raft/raftproto"
	"google.golang.org/grpc"
	"net"
	"time"
)

type Server struct {
	etc chan struct{}
}

func (s *Server) AppendEntry(ctx context.Context, req *raftproto.AppendEntryRequest) (rep *raftproto.AppendEntryResponse, err error) {
	fmt.Printf("%s: receive heartbeat\n", time.Now())
	s.etc <- struct{}{}
	fmt.Printf("%s: reset timeout signal sent\n", time.Now())
	return &raftproto.AppendEntryResponse{Term: 1, Success: true}, nil
}

func (s *Server) Run() {
	s.etc = make(chan struct{})
	for {
		select {
		case <-s.etc:
			fmt.Printf("%s: election timeout reset\n", time.Now())
			break
		case <-time.After(1 * time.Second):
			fmt.Printf("%s: election timeout triggered\n", time.Now())
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:22222")
	if err != nil {
		panic(err)
	}

	svr := Server{}
	go svr.Run()

	grpcServer := grpc.NewServer()
	raftproto.RegisterAppendEntryRPCServer(grpcServer, &svr)
	grpcServer.Serve(lis)
}
