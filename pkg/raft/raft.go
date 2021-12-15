package raft

import (
	"context"
	"fmt"
	transport "github.com/Jille/raft-grpc-transport"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	"github.com/stream-stack/store/pkg/config"
	"github.com/stream-stack/store/pkg/index"
	"github.com/stream-stack/store/pkg/snapshot"
	"github.com/stream-stack/store/pkg/wal"
	"google.golang.org/grpc"
	"math"
	"os"
	"path/filepath"
)

const stable = "stable.dat"

var Raft *raft.Raft
var RaftManager *transport.Manager

func StartRaft(ctx context.Context) error {
	c := raft.DefaultConfig()
	c.SnapshotThreshold = math.MaxUint64
	if len(RaftId) <= 0 {
		RaftId, _ = os.Hostname()
	}
	c.LocalID = raft.ServerID(RaftId)

	ss, err := snapshot.NewSnapshotStore(ctx)
	if err != nil {
		return err
	}

	stablePath := filepath.Join(config.DataDir, stable)
	empty := isEmptyDir(stablePath)
	sdb, err := boltdb.NewBoltStore(stablePath)
	if err != nil {
		return fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, stablePath, err)
	}

	RaftManager = transport.New(raft.ServerAddress(config.Address), []grpc.DialOption{grpc.WithInsecure()})
	Raft, err = raft.NewRaft(c, index.FSM, wal.LogStore, sdb, ss, RaftManager.Transport())
	if err != nil {
		return fmt.Errorf(`raft.NewRaft: %v`, err)
	}

	if Bootstrap && empty {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(RaftId),
					Address:  raft.ServerAddress(config.Address),
				},
			},
		}
		f := Raft.BootstrapCluster(cfg)
		if err := f.Error(); err != nil {
			return fmt.Errorf("raft.Raft.BootstrapCluster: %v", err)
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
