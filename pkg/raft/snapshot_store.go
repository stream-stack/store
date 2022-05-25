package raft

import (
	"context"
	"fmt"
	"github.com/hashicorp/raft"
)

func NewSnapshotStore(ctx context.Context) (raft.SnapshotStore, error) {
	fss, err := raft.NewFileSnapshotStore(snapshotDataDir, snapshotRetain, nil)
	if err != nil {
		return nil, fmt.Errorf(`raft.NewFileSnapshotStore(%q, ...): %v`, dataDir, err)
	}
	return fss, nil
}
