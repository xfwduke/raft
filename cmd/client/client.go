package main

import (
	"context"
	"github.com/xfwduke/raft/raftproto"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:22222", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	client := raftproto.NewAppendEntryRPCClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client.AppendEntry(
		ctx,
		&raftproto.AppendEntryRequest{
			Term:     1,
			LeaderId: 1,
		})
}
