package main

import (
	"context"
	"github.com/xfwduke/raft/raftproto"
	"google.golang.org/grpc"
	"time"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:22222", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	client := raftproto.NewAppendEntryRPCClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		client.AppendEntry(
			ctx,
			&raftproto.AppendEntryRequest{
				Term:     1,
				LeaderId: 1,
			})
		select {
		case <-time.After(600 * time.Millisecond):
			break
		}
	}

}
