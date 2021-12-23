package grpc

import (
	"context"
	"fmt"
	hraft "github.com/hashicorp/raft"
	"github.com/stream-stack/store/pkg/index"
	"github.com/stream-stack/store/pkg/protocol"
	"github.com/stream-stack/store/pkg/raft"
	"github.com/stream-stack/store/pkg/wal"
	"regexp"
	"strconv"
)

type EventService struct {
}

func (e *EventService) Subscribe(request *protocol.SubscribeRequest, server protocol.EventService_SubscribeServer) error {
	var compile *regexp.Regexp
	var err error
	if len(request.Regexp) != 0 {
		compile, err = regexp.Compile(request.Regexp)
		if err != nil {
			return err
		}
	}
	var start = request.Offset
	if start == 0 {
		start = 1
	}
	for {
		fmt.Println("当前offset:", start)
		//注册channel
		c := make(chan struct{})
		index.NotifyCh <- c
		fmt.Println("Subscribe.已注册channel")
		lastIndex, err := wal.LogStore.LastIndex()
		if err != nil {
			return err
		}
		fmt.Println("当前lastIndex:", lastIndex)
		for ; start <= lastIndex; start++ {
			select {
			case <-server.Context().Done():
				return server.Context().Err()
			default:
				if err := sendSubscribeResponse(start, server, compile); err != nil {
					return err
				}
			}
		}
		fmt.Println("赋值后,当前offset:", start)
		select {
		case <-server.Context().Done():
			return server.Context().Err()
		case <-c:
			fmt.Println("Subscribe.收到notify")
		}
	}
}

func sendSubscribeResponse(index uint64, server protocol.EventService_SubscribeServer, compile *regexp.Regexp) error {
	fmt.Println("当前正在发送Index:", index)
	//index = index + 1
	log := &hraft.Log{}
	err := wal.LogStore.GetLog(index, log)
	if err != nil {
		return err
	}
	if log.Type != hraft.LogCommand {
		return nil
	}
	if log.Data[0] == protocol.KeyValue {
		return nil
	}
	meta, err := protocol.ParseMeta(log.Extensions)
	if err != nil {
		return err
	}
	if compile != nil && !compile.MatchString(meta[0]) {
		return nil
	}
	fmt.Println("当前正在发送Index:", index)
	return server.Send(&protocol.ReadResponse{
		StreamName: meta[0],
		StreamId:   meta[1],
		EventId:    meta[2],
		Data:       log.Data[1:],
	})
}

func (e *EventService) Apply(ctx context.Context, request *protocol.ApplyRequest) (*protocol.ApplyResponse, error) {
	applyFuture := raft.Raft.ApplyLog(hraft.Log{
		Data:       protocol.AddApplyFlag(request.Data),
		Extensions: protocol.FormatApplyMeta(request.StreamName, request.StreamId, request.EventId),
	}, applyLogTimeout)
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

func (e *EventService) Read(ctx context.Context, request *protocol.ReadRequest) (*protocol.ReadResponse, error) {
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
	return &protocol.ReadResponse{
		StreamName: meta[0],
		StreamId:   meta[1],
		EventId:    meta[2],
		Data:       log.Data[1:],
	}, nil
}

func NewEventService() protocol.EventServiceServer {
	return &EventService{}
}
