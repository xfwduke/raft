package main

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/xfwduke/raft/rpcdemoproto"
	"google.golang.org/grpc"
	"time"
)

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	conn, err := grpc.DialContext(ctx, "127.0.0.1:10001", grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		//panic(err)
		log.Fatal(err)
	}

	hbClient := rpcdemoproto.NewHeartBeatServiceClient(conn)
	//vtClient := rpcdemoproto.NewVoteServiceClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var idx uint64 = 0
	for {
		hbrep, err := hbClient.KeepHearBeat(ctx, &rpcdemoproto.AppendEntriesRequest{Term: idx})
		if err != nil {
			panic(err)
		}
		log.Infof("recv AppendEntriesResponse: %v", hbrep)

		//vtrep, err := vtClient.Vote(ctx, &rpcdemoproto.VoteRequest{Term: 2 * idx})
		//if err != nil {
		//	panic(err)
		//}
		//log.Infof("recv VoteResponse: %v", vtrep)
		time.Sleep(2 * time.Second)
		idx++
	}
}
