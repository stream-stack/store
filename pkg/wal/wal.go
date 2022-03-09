package wal

import (
	"bytes"
	"context"
	"fmt"
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
	firstIndex, _ := l.log.FirstIndex()
	lastIndex, _ := l.log.LastIndex()

	read, err := l.log.Read(index)
	if err != nil {
		fmt.Println("l.log.Read 错误, index:", index, "firstIndex:", firstIndex, "lastIndex:", lastIndex, "err:", err)
		return err
	}
	err = decodeMsgPack(read, log)
	if err != nil {
		return err
	}
	return nil
}

func (l *LogStoreImpl) StoreLog(log *raft.Log) error {
	firstIndex, _ := l.log.FirstIndex()
	lastIndex, _ := l.log.LastIndex()
	fmt.Println("StoreLog, index:", log.Index, "firstIndex:", firstIndex, "lastIndex:", lastIndex)
	pack, err := encodeMsgPack(log)
	if err != nil {
		fmt.Println("StoreLog error", err)
		return err
	}
	return l.log.Write(log.Index, pack.Bytes())
}

func (l *LogStoreImpl) StoreLogs(logs []*raft.Log) error {
	firstIndex, _ := l.log.FirstIndex()
	lastIndex, _ := l.log.LastIndex()
	fmt.Println("StoreLogs", logs[0].Index, "firstIndex:", firstIndex, "lastIndex:", lastIndex)
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
	//err := l.log.TruncateFront(max - min)
	//if err != nil {
	//	return err
	//}
	err := l.log.TruncateBack(min - 1)
	if err != nil {
		return err
	}
	firstIndex, _ := l.log.FirstIndex()
	lastIndex, _ := l.log.LastIndex()
	fmt.Println("DeleteRange success", "first", firstIndex, "last", lastIndex, "delete args", min, max)
	return nil
}

const walDir = "wal"

func StartWalEngine(ctx context.Context) error {
	var walLogFormat wal.LogFormat
	if binaryLogFormat {
		walLogFormat = wal.Binary
	} else {
		walLogFormat = wal.JSON
	}
	open, err := wal.Open(filepath.Join(config.DataDir, walDir), &wal.Options{
		NoSync:           noSync,
		SegmentSize:      segmentSize,
		LogFormat:        walLogFormat,
		SegmentCacheSize: segmentCacheSize,
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
