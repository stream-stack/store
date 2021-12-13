package snapshot

import (
	"context"
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/stream-stack/store/pkg/config"
	"os"
)

func NewSnapshotStore(ctx context.Context) (raft.SnapshotStore, error) {
	fss, err := raft.NewFileSnapshotStore(config.DataDir, 3, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf(`raft.NewFileSnapshotStore(%q, ...): %v`, config.DataDir, err)
	}
	return fss, nil
}
