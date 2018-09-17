package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/xfwduke/raft/heartbeatdemoproto"
	"golang.org/x/net/context"
	"math/rand"
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
		rand.Seed(time.Now().UnixNano())
		heartBeatInterval := 100 + rand.Int31n(200)
		select {
		case <-hb.hbch:
			log.Info("timeout reset")
		case <-time.After(time.Duration(heartBeatInterval) * time.Millisecond):
			log.Info("timeout triggered")
			if autoTrans == true {
				return
			}
		}
	}
}
