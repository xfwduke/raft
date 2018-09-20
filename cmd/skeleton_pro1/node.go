package main

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"github.com/xfwduke/raft/rpcdemoproto"
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
	listenURL string
	peerNodes []string

	peerConnections []*grpc.ClientConn
	//heartBeatClients     []rpcdemoproto.HeartBeatServiceClient
	voteClients          []rpcdemoproto.VoteServiceClient
	appendEntriesClients []rpcdemoproto.AppendEntriesServiceClient

	state    NodeState
	stateMux sync.Mutex

	lis net.Listener

	grpcServer *grpc.Server
	grpcLis    net.Listener
	mux        cmux.CMux

	//heartBeatRestC    chan struct{}
	//heartBeatTriggerC chan struct{}
	//voteResetElectionTimeoutC          chan struct{}
	//appendEntriesResetElectionTimeoutC chan struct{}
	cancelElectionTimeout context.CancelFunc
	cancelElection        context.CancelFunc

	currentTerm uint64
	votedFor    uint64
	nodeId      uint64
}

func (nd *Node) AppendEntries(ctx context.Context, req *rpcdemoproto.AppendEntriesRequest) (*rpcdemoproto.AppendEntriesResponse, error) {
	nd.cancelElectionTimeout()
	nd.cancelElection()
	nd.transTo(Follower)

	if req.Term > nd.currentTerm { // and trans to follower
		nd.currentTerm = req.Term
	}
	if req.Log == nil || len(req.Log) == 0 { //HeartBeat

	} else {

	}
}

func (nd *Node) Vote(ctx context.Context, req *rpcdemoproto.VoteRequest) (*rpcdemoproto.VoteResponse, error) {
	nd.cancelElectionTimeout()

	if req.Term < nd.currentTerm {
		return &rpcdemoproto.VoteResponse{Term: nd.currentTerm, VoteGranted: false}, nil
	}
	if nd.votedFor == 0 || nd.votedFor == req.CandidateId {
		return &rpcdemoproto.VoteResponse{Term: req.Term, VoteGranted: true}, nil
	}
	return &rpcdemoproto.VoteResponse{Term: req.Term, VoteGranted: false}, nil
}

//func (nd *Node) KeepHearBeat(ctx context.Context, req *rpcdemoproto.AppendEntriesRequest) (*rpcdemoproto.AppendEntriesResponse, error) {
//	nd.heartBeatRestC <- struct{}{}
//	//ToDo
//	//process req
//	return &rpcdemoproto.AppendEntriesResponse{Term: req.Term}, nil
//}

func (nd *Node) Startup() {
	nd.bindLis()
	nd.connectPeers()
	go nd.startPRCService()

	etCtx, etCancel := context.WithCancel(context.Background())
	nd.cancelElectionTimeout = etCancel
	go nd.startElectionTimeout(etCtx)

	//for {
	//	switch nd.state {
	//	case Follower:
	//		nd.onAsFollower()
	//	case Candidate:
	//		nd.onAsCandidate()
	//	case Leader:
	//		nd.onAsLeader()
	//	}
	//}
}

//func (nd *Node) transTo(state NodeState) {
//	switch state {
//	case Follower:
//		nd.state = Follower
//		nd.onAsFollower()
//	case Candidate:
//		nd.state = Candidate
//		nd.onAsCandidate()
//	case Leader:
//		nd.state = Leader
//		nd.onAsLeader()
//	}
//}
func (nd *Node) bindLis() {
	lis, err := net.Listen("tcp", nd.listenURL)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("listen %s success", nd.listenURL)
	nd.lis = lis
}

func (nd *Node) connectPeers() {
	for _, peer := range nd.peerNodes {
		go func(peer string) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			conn, err := grpc.DialContext(ctx, peer, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Fatal("connect to %s failed: %s", peer, err)
			}
			log.Infof("connect to peer node %s success", peer)
			//nd.heartBeatClients = append(nd.heartBeatClients, rpcdemoproto.NewHeartBeatServiceClient(conn))
			nd.voteClients = append(nd.voteClients, rpcdemoproto.NewVoteServiceClient(conn))
			nd.appendEntriesClients = append(nd.appendEntriesClients, rpcdemoproto.NewAppendEntriesServiceClient(conn))
		}(peer)
	}
}

func (nd *Node) transTo(s NodeState) {
	nd.stateMux.Lock()
	defer nd.stateMux.Unlock()
	nd.state = s
}
func (nd *Node) getState() NodeState {
	nd.stateMux.Lock()
	defer nd.stateMux.Unlock()
	return nd.state
}

func (nd *Node) startPRCService() {
	nd.mux = cmux.New(nd.lis)
	nd.grpcLis = nd.mux.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	nd.grpcServer = grpc.NewServer()

	//rpcdemoproto.RegisterHeartBeatServiceServer(nd.grpcServer, nd)
	rpcdemoproto.RegisterVoteServiceServer(nd.grpcServer, nd)
	rpcdemoproto.RegisterAppendEntriesServiceServer(nd.grpcServer, nd)

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
}

func (nd *Node) startElectionTimeout(ctx context.Context) {
	eCtx, eCancel := context.WithCancel(context.Background())
	nd.cancelElection = eCancel

	for {
		rand.Seed(time.Now().UnixNano())
		electionTimeout := 150 + rand.Int31n(150)
		log.Infof("election timeout set to %d", electionTimeout)
		select {
		case <-ctx.Done():
			log.Infof("election timeout canceled")
		case <-time.After(time.Duration(electionTimeout) * time.Millisecond):
			log.Info("election timeout triggered")
			nd.transTo(Candidate)
			nd.doElection(eCtx)
		}
	}
}

func (nd *Node) onAsFollower() {
	log.Info("state = Follower")

}
func (nd *Node) doVote(ctx context.Context) {
	grantCount := 0
	for _, cli := range nd.voteClients {
		go func(cli rpcdemoproto.VoteServiceClient) {
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			resp, err := cli.Vote(
				ctx,
				&rpcdemoproto.VoteRequest{
					Term:        nd.currentTerm,
					CandidateId: nd.votedFor,
				})
			if err != nil {
				log.Errorf("error when send vote request: %s", err)
			}
			if resp.VoteGranted == true {
				grantCount += 1
			}
		}(cli)
	}
	if float64((grantCount+1.0)/(len(nd.peerNodes)+1.0)) > 0.5 {
		//win
	} else {
		//lose
	}
}

func (nd *Node) doElection(ctx context.Context) {
	go func() {
		defer nd.cancelElection()

		nd.currentTerm += 1
		nd.votedFor = nd.nodeId
		nd.cancelElectionTimeout()

	}()

	select {
	case <-ctx.Done():
		log.Info("election canceled or done")
	}

	//
	//grantCount := 0
	//for _, cli := range nd.voteClients {
	//	go func(cli rpcdemoproto.VoteServiceClient) {
	//		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	//
	//		defer cancel()
	//		resp, err := cli.Vote(
	//			ctx,
	//			&rpcdemoproto.VoteRequest{
	//				Term:        nd.currentTerm,
	//				CandidateId: nd.votedFor,
	//			})
	//		if err != nil {
	//			log.Errorf("error when send vote request: %s", err)
	//		}
	//		if resp.VoteGranted == true {
	//			grantCount += 1
	//		}
	//	}(cli)
	//}
	//if float64((grantCount+1.0)/(len(nd.peerNodes)+1.0)) > 0.5 {
	//	//win
	//} else {
	//	//lose
	//}

}

func (nd *Node) onAsCandidate() {
	//go nd.doElection()
	//select {
	//case <-nd.heartBeatTriggerC:
	//	break
	//}
	//nd.currentTerm += 1
	//nd.votedFor = nd.nodeId
	//nd.heartBeatRestC <- struct{}{}
	//
	//for _, cli := range nd.voteClients {
	//	go func(cli rpcdemoproto.VoteServiceClient) {
	//		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	//		defer cancel()
	//		resp, err := cli.Vote(
	//			ctx,
	//			&rpcdemoproto.VoteRequest{
	//				Term:        nd.currentTerm,
	//				CandidateId: nd.votedFor,
	//			})
	//		if err != nil {
	//			log.Errorf("error when send vote request: %s", err)
	//		}
	//	}(cli)
	//}
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
}

func newNode(listenURL string, peerNodes []string) (*Node, error) {
	hashAddr, err := encodeAddr(listenURL)
	if err != nil {
		return nil, err
	}
	return &Node{
		listenURL:   listenURL,
		peerNodes:   peerNodes,
		state:       Follower,
		currentTerm: 0,
		nodeId:      hashAddr,
	}, nil
}
