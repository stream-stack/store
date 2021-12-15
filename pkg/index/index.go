package index

import (
	"context"
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/stream-stack/store/pkg/config"
	"github.com/stream-stack/store/pkg/protocol"
	"github.com/syndtr/goleveldb/leveldb"
	"io"
	"path/filepath"
)

const dbName = "index"

var FSM raft.FSM
var KVDb *leveldb.DB

type FSMImpl struct {
}

/*
三种情况:
1.producer,储存 聚合id+eventId,log.index
2.consumer,储存 订阅者id,offset
3.read,获取 聚合id+eventId,log.index
*/
func (f *FSMImpl) Apply(log *raft.Log) interface{} {
	//TODO:实现构建快照
	fmt.Println("Apply:")
	fmt.Printf("%+v", log)
	data := log.Data
	flag := data[0]
	switch flag {
	case protocol.Apply:
		//TODO:error 处理?
		return KVDb.Put(log.Extensions, protocol.Uint64ToBytes(log.Index), nil)
	case protocol.Offset:
		return KVDb.Put(log.Extensions, log.Data[1:], nil)
	}
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
	var err error
	KVDb, err = leveldb.OpenFile(filepath.Join(config.DataDir, dbName), nil)
	if err != nil {
		return err
	}
	go func() {
		select {
		case <-ctx.Done():
			KVDb.Close()
		}
	}()
	FSM = &FSMImpl{}
	return nil
}
