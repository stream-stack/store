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
	"path/filepath"
)

const stable = "stable.dat"

var Raft *raft.Raft
var RaftManager *transport.Manager

func StartRaft(ctx context.Context) error {
	//_, port, err := net.SplitHostPort(Address)
	//if err != nil {
	//	logrus.Fatalf("failed to parse local address (%q): %v", Address, err)
	//}
	//sock, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	//if err != nil {
	//	logrus.Fatalf("failed to listen: %v", err)
	//}

	c := raft.DefaultConfig()
	c.SnapshotThreshold = math.MaxUint64
	c.LocalID = raft.ServerID(RaftId)

	ss, err := snapshot.NewSnapshotStore(ctx)
	if err != nil {
		return err
	}

	stablePath := filepath.Join(config.DataDir, stable)
	sdb, err := boltdb.NewBoltStore(stablePath)
	if err != nil {
		return fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, stablePath, err)
	}

	RaftManager = transport.New(raft.ServerAddress(config.Address), []grpc.DialOption{grpc.WithInsecure()})
	r, err := raft.NewRaft(c, index.FSM, wal.LogStore, sdb, ss, RaftManager.Transport())
	if err != nil {
		return fmt.Errorf(`raft.NewRaft: %v`, err)
	}

	if Bootstrap {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(RaftId),
					Address:  raft.ServerAddress(config.Address),
				},
			},
		}
		f := r.BootstrapCluster(cfg)
		if err := f.Error(); err != nil {
			return fmt.Errorf("raft.Raft.BootstrapCluster: %v", err)
		}
	}
	Raft = r
	return nil
}
