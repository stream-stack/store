package grpc

import (
	"context"
	"fmt"
	hraft "github.com/hashicorp/raft"
	"github.com/stream-stack/store/pkg/index"
	"github.com/stream-stack/store/pkg/protocol"
	"github.com/stream-stack/store/pkg/raft"
	"github.com/stream-stack/store/pkg/wal"
	"strconv"
	"time"
)

type EventService struct {
}

func (e *EventService) Offset(ctx context.Context, request *protocol.OffsetRequest) (*protocol.OffsetResponse, error) {
	applyFuture := raft.Raft.ApplyLog(hraft.Log{
		Data:       protocol.AddOffsetFlag(request.Offset),
		Extensions: []byte(request.SubscribeId),
	}, time.Second)
	//TODO:超时设置
	if err := applyFuture.Error(); err != nil {
		return &protocol.OffsetResponse{
			Ack:     true,
			Message: err.Error(),
		}, err
	}
	return &protocol.OffsetResponse{
		Ack:     true,
		Message: strconv.FormatUint(applyFuture.Index(), 10),
	}, nil
}

func (e *EventService) Apply(ctx context.Context, request *protocol.ApplyRequest) (*protocol.ApplyResponse, error) {
	fmt.Println("Apply=================")
	applyFuture := raft.Raft.ApplyLog(hraft.Log{
		Data:       protocol.AddApplyFlag(request.Data),
		Extensions: protocol.FormatApplyMeta(request.StreamName, request.StreamId, request.EventId),
	}, time.Second)
	//TODO:超时设置
	if err := applyFuture.Error(); err != nil {
		return &protocol.ApplyResponse{
			Ack:     true,
			Message: err.Error(),
		}, err
	}
	return &protocol.ApplyResponse{
		Ack:     true,
		Message: strconv.FormatUint(applyFuture.Index(), 10),
	}, nil
}

func (e *EventService) Get(ctx context.Context, request *protocol.GetRequest) (*protocol.GetResponse, error) {
	key := protocol.FormatApplyMeta(request.StreamName, request.StreamId, request.EventId)
	get, err := index.KVDb.Get(key, nil)
	if err != nil {
		return nil, err
	}
	dataIndex := protocol.BytesToUint64(get)
	log := &hraft.Log{}
	err = wal.LogStore.GetLog(dataIndex, log)
	if err != nil {
		return nil, err
	}
	meta, err := protocol.ParseMeta(log.Extensions)
	if err != nil {
		return nil, err
	}
	return &protocol.GetResponse{
		StreamName: meta[0],
		StreamId:   meta[1],
		EventId:    meta[2],
		Data:       log.Data[1:],
	}, nil
}

func NewEventService() protocol.EventServiceServer {
	return &EventService{}
}
