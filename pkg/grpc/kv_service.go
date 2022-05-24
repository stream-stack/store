package grpc

import (
	"context"
	"fmt"
	hraft "github.com/hashicorp/raft"
	protocol "github.com/stream-stack/common/protocol/store"
	"github.com/stream-stack/store/pkg/raft"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type KVService struct {
}

func (K *KVService) GetRange(ctx context.Context, request *protocol.GetRequest) (*protocol.GetRangeResponse, error) {
	return nil, nil
}

func (K *KVService) Put(ctx context.Context, request *protocol.PutRequest) (*protocol.PutResponse, error) {
	applyFuture := raft.Raft.ApplyLog(hraft.Log{
		Data:       request.Val,
		Extensions: request.Key,
	}, applyLogTimeout)
	if err := applyFuture.Error(); err != nil {
		return &protocol.PutResponse{
			Ack:     true,
			Message: err.Error(),
		}, err
	}
	return &protocol.PutResponse{
		Ack: true,
	}, nil
}

func (K *KVService) Get(ctx context.Context, request *protocol.GetRequest) (*protocol.GetResponse, error) {
	data, _, err := raft.GetLogByKey(request.Key)
	if err != nil {
		if err == raft.ErrKeyNotFound {
			return &protocol.GetResponse{}, status.Error(codes.NotFound, fmt.Sprintf("key %s not found", request.Key))
		}
		return &protocol.GetResponse{}, status.Error(codes.Unknown, err.Error())
	}
	return &protocol.GetResponse{Data: data}, nil
}

func NewKVService() protocol.KVServiceServer {
	return &KVService{}
}
