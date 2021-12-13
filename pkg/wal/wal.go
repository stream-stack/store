package wal

import (
	"bytes"
	"context"
	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/raft"
	"github.com/stream-stack/store/pkg/config"
	"github.com/tidwall/wal"
	"path/filepath"
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
	err = decodeMsgPack(read, log)
	if err != nil {
		return err
	}
	return nil
}

// Decode reverses the encode operation on a byte slice input
func decodeMsgPack(buf []byte, out interface{}) error {
	r := bytes.NewBuffer(buf)
	hd := codec.MsgpackHandle{}
	dec := codec.NewDecoder(r, &hd)
	return dec.Decode(out)
}

// Encode writes an encoded object to a new bytes buffer
func encodeMsgPack(in interface{}) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	hd := codec.MsgpackHandle{}
	enc := codec.NewEncoder(buf, &hd)
	err := enc.Encode(in)
	return buf, err
}

func (l *LogStoreImpl) StoreLog(log *raft.Log) error {
	pack, err := encodeMsgPack(log)
	if err != nil {
		return err
	}
	return l.log.Write(log.Index, pack.Bytes())
}

func (l *LogStoreImpl) StoreLogs(logs []*raft.Log) error {
	batch := new(wal.Batch)
	for _, log := range logs {
		pack, err := encodeMsgPack(log)
		if err != nil {
			return err
		}
		batch.Write(log.Index, pack.Bytes())
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

const walDir = "wal"

func StartWalEngine(ctx context.Context) error {
	open, err := wal.Open(filepath.Join(config.DataDir, walDir), &wal.Options{
		NoSync:           false,
		SegmentSize:      0,
		LogFormat:        wal.JSON,
		SegmentCacheSize: 0,
		DirPerms:         0,
		FilePerms:        0,
	})
	if err != nil {
		return err
	}
	go func() {
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
