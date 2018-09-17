package main

import (
	"context"
	"fmt"
	"github.com/xfwduke/raft/heartbeatdemoproto"
	"google.golang.org/grpc"
	"math/rand"
	"time"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:22222", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	client := demoproto.NewHeartBeatClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		rand.Seed(time.Now().UnixNano())
		heartbeat := 1500 + rand.Int31n(1000)
		_, err := client.KeepHeartBeat(
			ctx,
			&demoproto.HeartBeatRequest{})
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s: heartbeat sent\n", time.Now())
		select {
		case <-time.After(time.Duration(heartbeat) * time.Millisecond):
			break
		}
	}

}
