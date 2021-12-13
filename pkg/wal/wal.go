package wal

import (
	"context"
	"github.com/hashicorp/raft"
	"github.com/stream-stack/store/pkg/config"
	"github.com/tidwall/wal"
)

var LogStore raft.LogStore

type LogStoreImpl struct {
	log *wal.Log
}

func (l *LogStoreImpl) FirstIndex() (uint64, error) {
	return l.log.FirstIndex()
}

func (l *LogStoreImpl) LastIndex() (uint64, error) {
	return l.log.LastIndex()
}

func (l *LogStoreImpl) GetLog(index uint64, log *raft.Log) error {
	read, err := l.log.Read(index)
	if err != nil {
		return err
	}
	log.Data = read
	return nil
}

func (l *LogStoreImpl) StoreLog(log *raft.Log) error {
	return l.log.Write(log.Index, log.Data)
}

func (l *LogStoreImpl) StoreLogs(logs []*raft.Log) error {
	batch := new(wal.Batch)
	for i, log := range logs {
		batch.Write(uint64(i), log.Data)
	}
	return l.log.WriteBatch(batch)
}

func (l *LogStoreImpl) DeleteRange(min, max uint64) error {
	err := l.log.TruncateFront(min)
	if err != nil {
		return err
	}
	err = l.log.TruncateBack(max)
	if err != nil {
		return err
	}
	return nil
}

func StartWalEngine(ctx context.Context) error {
	open, err := wal.Open(config.DataDir, nil)
	if err != nil {
		return err
	}
	defer func() {
		select {
		case <-ctx.Done():
			open.Close()
		}
	}()

	LogStore = &LogStoreImpl{
		log: open,
	}
	return nil
}
