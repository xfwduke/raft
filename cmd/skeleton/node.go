package main

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"github.com/xfwduke/raft/rpcdemoproto"
	"google.golang.org/grpc"
	"math/rand"
	"net"
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
	listenURL string
	peerNodes []string

	peerConnections      []*grpc.ClientConn
	heartBeatClients     []rpcdemoproto.HeartBeatServiceClient
	voteClients          []rpcdemoproto.VoteServiceClient
	appendEntriesClients []rpcdemoproto.AppendEntriesServiceClient

	state NodeState

	grpcServer *grpc.Server
	grpcLis    net.Listener
	mux        cmux.CMux

	heartBeatC chan struct{}
}

func (nd *Node) AppendEntries(context.Context, *rpcdemoproto.AppendEntriesRequest) (*rpcdemoproto.AppendEntriesResponse, error) {
	panic("implement me")
}

func (nd *Node) Vote(context.Context, *rpcdemoproto.VoteRequest) (*rpcdemoproto.VoteResponse, error) {
	panic("implement me")
}

func (nd *Node) KeepHearBeat(ctx context.Context, req *rpcdemoproto.AppendEntriesRequest) (*rpcdemoproto.AppendEntriesResponse, error) {
	nd.heartBeatC <- struct{}{}
	//ToDo
	//process req
	return &rpcdemoproto.AppendEntriesResponse{Term: req.Term}, nil
}

func (nd *Node) Startup() {
	lis, err := net.Listen("tcp", nd.listenURL)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("listen %s success", nd.listenURL)

	nd.mux = cmux.New(lis)
	nd.grpcLis = nd.mux.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	nd.grpcServer = grpc.NewServer()

	rpcdemoproto.RegisterHeartBeatServiceServer(nd.grpcServer, nd)
	rpcdemoproto.RegisterVoteServiceServer(nd.grpcServer, nd)
	rpcdemoproto.RegisterAppendEntriesServiceServer(nd.grpcServer, nd)

	for _, peerURL := range nd.peerNodes {
		go func(pu string) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			conn, err := grpc.DialContext(ctx, peerURL, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Fatalf("connect to peer node %s failed: %s", peerURL, err)
			}
			log.Infof("connect to peer node %s success", peerURL)
			nd.heartBeatClients = append(nd.heartBeatClients, rpcdemoproto.NewHeartBeatServiceClient(conn))
			nd.voteClients = append(nd.voteClients, rpcdemoproto.NewVoteServiceClient(conn))
			nd.appendEntriesClients = append(nd.appendEntriesClients, rpcdemoproto.NewAppendEntriesServiceClient(conn))
		}(peerURL)
	}
	nd.transTo(Follower)
}

func (nd *Node) transTo(state NodeState) {
	switch state {
	case Follower:
		nd.state = Follower
		nd.onAsFollower()
	case Candidate:
		nd.state = Candidate
		nd.onAsCandidate()
	case Leader:
		nd.state = Leader
		nd.onAsLeader()
	}
}

func (nd *Node) onAsFollower() {
	log.Info("state = Follower")

	go func() {
		go func() {
			err := nd.grpcServer.Serve(nd.grpcLis)
			if err != nil {
				log.Fatal(err)
			}
		}()
		err := nd.mux.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	for {
		rand.Seed(time.Now().UnixNano())
		electionTimeout := 150 + rand.Int31n(150)
		log.Infof("set election timeout to: %d", electionTimeout)
		select {
		case <-nd.heartBeatC:
			log.Info("election timeout reset")
		case <-time.After(time.Duration(electionTimeout) * time.Millisecond):
			log.Info("election timeout triggered")
			//ToDo trans to Candidate
		}
	}

	nd.transTo(Candidate)
}

func (nd *Node) onAsCandidate() {
	nd.transTo(Leader)
}

func (nd *Node) onAsLeader() {
	for _, heartBeatClient := range nd.heartBeatClients {
		go func(cli rpcdemoproto.HeartBeatServiceClient) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			for {
				response, err := cli.KeepHearBeat(ctx, &rpcdemoproto.AppendEntriesRequest{})
				if err != nil {
					//ToDo
					//shall we auto-reconnect?
					log.Error(err)
				}
				//ToDo
				log.Info(response)
			}
		}(heartBeatClient)
	}
	nd.transTo(Follower)
}

func newNode(listenURL string, peerNodes []string) *Node {
	return &Node{
		listenURL:  listenURL,
		peerNodes:  peerNodes,
		state:      Follower,
		heartBeatC: make(chan struct{}),
	}
}


