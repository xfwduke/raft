package main

import (
	"context"
	"fmt"
	"github.com/xfwduke/raft/raftproto"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
}

func (s *Server) AppendEntry(context.Context, *raftproto.AppendEntryRequest) (*raftproto.AppendEntryResponse, error) {
	fmt.Println("AppendEntries received")
	return &raftproto.AppendEntryResponse{Term: 1, Success: true}, nil
}

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:22222")
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	raftproto.RegisterAppendEntryRPCServer(grpcServer, &Server{})
	grpcServer.Serve(lis)
}

