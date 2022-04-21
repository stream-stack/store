package grpc

import (
	"context"
	"fmt"
	hraft "github.com/hashicorp/raft"
	protocol "github.com/stream-stack/common/protocol/store"
	"github.com/stream-stack/store/pkg/index"
	"github.com/stream-stack/store/pkg/raft"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
)

type KVService struct {
}

func (K *KVService) GetRange(ctx context.Context, request *protocol.GetRequest) (*protocol.GetRangeResponse, error) {
	iterator := index.KVDb.NewIterator(util.BytesPrefix(request.Key), nil)
	defer iterator.Release()
	responses := make([]*protocol.GetResponse, 0)
	for iterator.Next() {
		responses = append(responses, &protocol.GetResponse{Data: iterator.Value()})
	}
	return &protocol.GetRangeResponse{Response: responses}, nil
}

func (K *KVService) Put(ctx context.Context, request *protocol.PutRequest) (*protocol.PutResponse, error) {
	applyFuture := raft.Raft.ApplyLog(hraft.Log{
		Data:       protocol.AddKeyValueFlag(request.Val),
		Extensions: request.Key,
	}, applyLogTimeout)
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
		if err == errors.ErrNotFound {
			return &protocol.GetResponse{}, status.Error(codes.NotFound, fmt.Sprintf("key %s not found", request.Key))
		}
		return &protocol.GetResponse{}, status.Error(codes.Unknown, err.Error())
	}
	return &protocol.GetResponse{Data: get}, nil
}

func NewKVService() protocol.KVServiceServer {
	return &KVService{}
}
