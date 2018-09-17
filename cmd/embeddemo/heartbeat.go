package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/xfwduke/raft/heartbeatdemoproto"
	"golang.org/x/net/context"
	"time"
)

type HeartBeatServer struct {
	hbch chan struct{}
}

func (hb *HeartBeatServer) KeepHeartBeat(context.Context, *demoproto.HeartBeatRequest) (*demoproto.HeartBeatResponse, error) {
	log.Info("heartbeat request received")
	hb.hbch <- struct{}{}
	log.Info("timeout reset signal sent")
	return &demoproto.HeartBeatResponse{}, nil
}

func (hb *HeartBeatServer) StartTimeoutLoop() {
	for {
		select {
		case <-hb.hbch:
			log.Info("timeout reset")
		case <-time.After(2 * time.Second):
			log.Info("timeout triggered")
		}
	}
}

