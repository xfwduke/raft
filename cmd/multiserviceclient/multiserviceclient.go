package main

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/xfwduke/raft/rpcdemoproto"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:22222", grpc.WithBlock())
	if err != nil {
		panic(err)
	}

	hbClient := rpcdemoproto.NewHeartBeatServiceClient(conn)
	vtClient := rpcdemoproto.NewVoteServiceClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		hbrep, err := hbClient.KeepHearBeat(ctx, &rpcdemoproto.AppendEntriesRequest{})
		if err != nil {
			panic(err)
		}
		log.Infof("recv AppendEntriesResponse: %v", hbrep)

		vtrep, err := vtClient.Vote(ctx, &rpcdemoproto.VoteRequest{})
		if err != nil {
			panic(err)
		}
		log.Infof("recv VoteResponse: %v", vtrep)
	}
}
