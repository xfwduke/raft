package main

import (
	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xfwduke/raft/heartbeatdemoproto"
	"google.golang.org/grpc"
	"os"
	"time"
)

var rootCmd = &cobra.Command{}
var listenURL string
var peerNodeURLs []string

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)

	rootCmd.PersistentFlags().StringVar(&listenURL, "listen", "", "")
	rootCmd.PersistentFlags().StringArrayVar(&peerNodeURLs, "peer-nodes", nil, "")
	rootCmd.AddCommand(&cobra.Command{
		Use:   "follower",
		Short: "run as follower",
		Run:   RunAsFollower,
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "leader",
		Short: "run as leader",
		Run:   RunAsLeader,
	})
}

func RunAsFollower(cmd *cobra.Command, args []string) {
	node, err := NewNode()
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	demoproto.RegisterHeartBeatServer(grpcServer, node.HBServer)
	go grpcServer.Serve(node.Lis)
	node.HBServer.StartTimeoutLoop()
}

func RunAsLeader(cmd *cobra.Command, args []string) {
	node, err := NewNode()
	if err != nil {
		panic(err)
	}

	for _, peerNodeURL := range peerNodeURLs {
		ctx, _ := context.WithTimeout(context.Background(), 200*time.Millisecond)
		conn, err := grpc.DialContext(ctx, peerNodeURL, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			panic(err)
		}
		log.Infof("connect to follower %s success", peerNodeURL)
		node.HBClients = append(node.HBClients, demoproto.NewHeartBeatClient(conn))
	}

	node.StartSendHeartBeat()
}

func main() {
	rootCmd.Execute()
}
