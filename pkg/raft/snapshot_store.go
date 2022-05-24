package raft

import (
	"context"
	"fmt"
	"github.com/hashicorp/raft"
	"os"
)

func NewSnapshotStore(ctx context.Context) (raft.SnapshotStore, error) {
	fss, err := raft.NewFileSnapshotStore(dataDir, 3, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf(`raft.NewFileSnapshotStore(%q, ...): %v`, dataDir, err)
	}
	return fss, nil
}
