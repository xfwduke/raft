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

type NodeState int

const (
	Follower NodeState = iota
	Candidate
	Leader
)

func (ns NodeState) String() string {
	return [...]string{"Follower", "Candidate", "Leader"}[ns]
}

type Node struct {
	Lis       net.Listener
	HBServer  *HeartBeatServer
	HBClients []demoproto.HeartBeatClient
	State     NodeState
}

func (nd *Node) TransTo(state NodeState) {
	switch state {
	case Follower:
		nd.State = Follower
		nd.onAsFollower()
	case Candidate:
		nd.State = Candidate
		nd.onAsFollower()
	case Leader:
		nd.State = Leader
		nd.onAsLeader()
	}
}

func (nd *Node) onAsFollower() {
	log.Info("state = Follower")
	nd.HBServer.StartTimeoutLoop()
	nd.TransTo(Candidate)
}

func (nd *Node) onAsCandidate() {
	log.Info("state = Candidate")
	nd.TransTo(Leader)
}

func (nd *Node) onAsLeader() {
	log.Info("state = Leader")
	ctx, cancel := context.WithCancel(context.Background())
	nd.startSendHeartBeat(ctx)

	select {
	case <-time.After(5 * time.Second):
		cancel()
		break
	}
	nd.TransTo(Follower)
}

func (nd *Node) startSendHeartBeat(ctx context.Context) {
	var wg sync.WaitGroup
	for _, cli := range nd.HBClients {
		wg.Add(1)
		go func(cli demoproto.HeartBeatClient) {
			dtx, cancel := context.WithCancel(context.Background())
			defer cancel()
			for {
				rand.Seed(time.Now().UnixNano())
				electionTimeout := 150 + rand.Int31n(200)
				_, err := cli.KeepHeartBeat(dtx, &demoproto.HeartBeatRequest{})
				if err != nil {
					panic(err)
				}
				log.Infof("heartbeat sent to %v", cli)
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Duration(electionTimeout) * time.Millisecond):
					break
				}
			}
			wg.Done()
		}(cli)
	}
	wg.Wait()
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
