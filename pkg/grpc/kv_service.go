package grpc

import (
	"context"
	hraft "github.com/hashicorp/raft"
	"github.com/stream-stack/store/pkg/index"
	"github.com/stream-stack/store/pkg/protocol"
	"github.com/stream-stack/store/pkg/raft"
	"strconv"
	"time"
)

type KVService struct {
}

func (K *KVService) Put(ctx context.Context, request *protocol.PutRequest) (*protocol.PutResponse, error) {
	applyFuture := raft.Raft.ApplyLog(hraft.Log{
		Data:       protocol.AddKeyValueFlag(request.Val),
		Extensions: request.Key,
	}, time.Second)
	//TODO:超时设置
	if err := applyFuture.Error(); err != nil {
		return &protocol.PutResponse{
			Ack:     true,
			Message: err.Error(),
		}, err
	}
	return &protocol.PutResponse{
		Ack:     true,
		Message: strconv.FormatUint(applyFuture.Index(), 10),
	}, nil
}

func (K *KVService) Get(ctx context.Context, request *protocol.GetRequest) (*protocol.GetResponse, error) {
	get, err := index.KVDb.Get(request.Key, nil)
	if err != nil {
		return &protocol.GetResponse{}, err
	}
	return &protocol.GetResponse{Data: get}, nil
}

func NewKVService() protocol.KVServiceServer {
	return &KVService{}
}
