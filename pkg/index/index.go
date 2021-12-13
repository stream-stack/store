package index

import (
	"context"
	"github.com/hashicorp/raft"
	"github.com/stream-stack/store/pkg/config"
	"github.com/syndtr/goleveldb/leveldb"
	"io"
	"path/filepath"
)

const dbName = "index.db"

var FSM raft.FSM

type FSMImpl struct {
	db *leveldb.DB
}

func (f *FSMImpl) Apply(log *raft.Log) interface{} {
	//TODO:实现构建快照
	return nil
}

func (f *FSMImpl) Snapshot() (raft.FSMSnapshot, error) {
	return &FSMSnapshotImpl{}, nil
}

func (f *FSMImpl) Restore(closer io.ReadCloser) error {
	return nil
}

type FSMSnapshotImpl struct {
}

func (F *FSMSnapshotImpl) Persist(sink raft.SnapshotSink) error {
	return nil
}

func (F *FSMSnapshotImpl) Release() {

}

func StartFSM(ctx context.Context) error {
	db, err := leveldb.OpenFile(filepath.Join(config.DataDir, dbName), nil)
	if err != nil {
		return err
	}
	defer func() {
		select {
		case <-ctx.Done():
			db.Close()
		}
	}()
	FSM = &FSMImpl{db: db}
	return nil
}
