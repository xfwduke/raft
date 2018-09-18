package main

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/xfwduke/raft/heartbeatdemoproto"
	"google.golang.org/grpc"
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
	Lis        net.Listener
	grpcServer *grpc.Server
	HBServer   *HeartBeatServer
	HBClients  []demoproto.HeartBeatClient
	State      NodeState
}

func (nd *Node) Startup() {
	nd.grpcServer = grpc.NewServer()
	demoproto.RegisterHeartBeatServer(nd.grpcServer, nd.HBServer)
	nd.TransTo(Follower)
}

func (nd *Node) TransTo(state NodeState) {
	switch state {
	case Follower:
		nd.State = Follower
		nd.onAsFollower()
	case Candidate:
		nd.State = Candidate
		nd.onAsCandidate()
	case Leader:
		nd.State = Leader
		nd.onAsLeader()
	}
}

func (nd *Node) onAsFollower() {
	log.Info("state = Follower")
	go nd.grpcServer.Serve(nd.Lis)
	nd.HBServer.StartTimeoutLoop()
	nd.grpcServer.GracefulStop()
	if autoTrans == true {
		time.Sleep(5 * time.Second)
	}
	nd.TransTo(Candidate)
}

func (nd *Node) onAsCandidate() {
	log.Info("state = Candidate")
	nd.TransTo(Leader)
}

func (nd *Node) onAsLeader() {
	log.Info("state = Leader")
	for _, peerNodeURL := range peerNodeURLs {
		dtx, _ := context.WithTimeout(context.Background(), 200*time.Millisecond)
		conn, err := grpc.DialContext(dtx, peerNodeURL, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			panic(err)
		}
		log.Infof("connect to follower %s success", peerNodeURL)
		nd.HBClients = append(nd.HBClients, demoproto.NewHeartBeatClient(conn))
	}

	ctx, cancel := context.WithCancel(context.Background())
	go nd.startSendHeartBeat(ctx)

	select {
	case <-time.After(5 * time.Second):
		cancel()
		log.Info("trans from leader to follower triggered")
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
			defer wg.Done()
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
