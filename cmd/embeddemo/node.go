package main

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/xfwduke/raft/heartbeatdemoproto"
	"math/rand"
	"net"
	"sync"
	"time"
)

type Node struct {
	Lis       net.Listener
	HBServer  *HeartBeatServer
	HBClients []demoproto.HeartBeatClient
}

func (nd *Node) StartSendHeartBeat() {
	var wg sync.WaitGroup
	for _, cli := range nd.HBClients {
		wg.Add(1)
		go func(cli demoproto.HeartBeatClient) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			for {
				rand.Seed(time.Now().UnixNano())
				hbInterval := 1500 + rand.Int31n(1000)
				log.Infof("heartbeat interval = %d", hbInterval)
				_, err := cli.KeepHeartBeat(ctx, &demoproto.HeartBeatRequest{})
				if err != nil {
					panic(err)
				}
				log.Infof("heartbeat sent to %v", cli)
				select {
				case <-time.After(time.Duration(hbInterval) * time.Millisecond):
					break
				}
			}
			wg.Done()
		}(cli)
	}
	wg.Wait()
}

func NewNode() (*Node, error) {
	lis, err := net.Listen("tcp", listenURL)

	return &Node{
		Lis: lis,
		HBServer: &HeartBeatServer{
			hbch: make(chan struct{}),
		},
	}, err
}
