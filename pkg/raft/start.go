package raft

import (
	"context"
	"fmt"
	transport "github.com/Jille/raft-grpc-transport"
	"github.com/hashicorp/raft"
	"github.com/sirupsen/logrus"
	"github.com/stream-stack/store/pkg/config"
	"google.golang.org/grpc"
	"math"
	"os"
)

var Raft *raft.Raft
var RaftManager *transport.Manager
var Store *BadgerStore

func StartRaft(ctx context.Context) error {
	c := raft.DefaultConfig()

	//TODO:配置
	//TODO:快照配置

	c.SnapshotThreshold = math.MaxUint64
	if len(raftId) <= 0 {
		raftId, _ = os.Hostname()
	}
	c.LocalID = raft.ServerID(raftId)
	empty := isEmptyDir(dataDir)

	ss, err := NewSnapshotStore(ctx)
	if err != nil {
		return err
	}

	logrus.Debugf("dataDir:%s,empty:%v", dataDir, empty)
	Store, err = NewBadgerStore(dataDir)
	if err != nil {
		logrus.Debugf("new Badger Store error:%v", err)
		return err
	}

	RaftManager = transport.New(raft.ServerAddress(config.Address), []grpc.DialOption{grpc.WithInsecure()})
	Raft, err = raft.NewRaft(c, &fsmImpl{}, Store, Store, ss, RaftManager.Transport())
	if err != nil {
		return fmt.Errorf(`new raft instance error: %v`, err)
	}

	if bootstrap && empty {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(raftId),
					Address:  raft.ServerAddress(config.Address),
				},
			},
		}
		if err := Raft.BootstrapCluster(cfg).Error(); err != nil {
			return fmt.Errorf("bootstarp raft instance error: %v", err)
		}
	}
	return nil
}

func isEmptyDir(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return true
		}
	}
	return false
}
