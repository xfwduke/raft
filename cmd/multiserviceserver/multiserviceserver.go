package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"github.com/xfwduke/raft/rpcdemoproto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
)

type HBSV struct {
}

func (*HBSV) KeepHearBeat(ctx context.Context, req *rpcdemoproto.AppendEntriesRequest) (*rpcdemoproto.AppendEntriesResponse, error) {
	log.Infof("recv AppendEntriesRequest: %v", req)
	return &rpcdemoproto.AppendEntriesResponse{}, nil
}

type VTSV struct {
}

func (*VTSV) Vote(ctx context.Context, req *rpcdemoproto.VoteRequest) (*rpcdemoproto.VoteResponse, error) {
	log.Infof("recv VoteRequest: %v", req)
	return &rpcdemoproto.VoteResponse{}, nil
}

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:22222")
	if err != nil {
		panic(err)
	}

	m := cmux.New(lis)
	grpcL := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))

	grpcServer := grpc.NewServer()
	rpcdemoproto.RegisterHeartBeatServiceServer(grpcServer, &HBSV{})
	rpcdemoproto.RegisterVoteServiceServer(grpcServer, &VTSV{})

	go grpcServer.Serve(grpcL)
	m.Serve()
}
