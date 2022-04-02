package grpc

import (
	"context"
	"fmt"
	"github.com/Knetic/govaluate"
	hraft "github.com/hashicorp/raft"
	"github.com/sirupsen/logrus"
	"github.com/stream-stack/store/pkg/index"
	"github.com/stream-stack/store/pkg/protocol"
	"github.com/stream-stack/store/pkg/raft"
	"github.com/stream-stack/store/pkg/wal"
	"strconv"
)

type EventService struct {
}

func (e *EventService) Subscribe(request *protocol.SubscribeRequest, server protocol.EventService_SubscribeServer) error {
	var expression *govaluate.EvaluableExpression
	var err error
	if len(request.Regexp) != 0 {
		expression, err = govaluate.NewEvaluableExpression(request.Regexp)
		if err != nil {
			return err
		}
	}
	var start = request.Offset
	if start == 0 {
		start = 1
	}
	for {
		logrus.Debugf("当前offset:%d", start)
		//注册channel
		c := make(chan struct{})
		index.NotifyCh <- c
		logrus.Debugf("Subscribe.已注册channel")
		lastIndex, err := wal.LogStore.LastIndex()
		if err != nil {
			logrus.Errorf("获取LastIndex error:%v", err)
			return err
		}
		logrus.Debugf("当前lastIndex:%d", lastIndex)
		for ; start <= lastIndex; start++ {
			select {
			case <-server.Context().Done():
				return server.Context().Err()
			default:
				if err := sendSubscribeResponse(start, server, expression); err != nil {
					logrus.Errorf("sendSubscribeResponse error:%v", err)
					return err
				}
			}
		}
		logrus.Debugf("赋值后,当前offset:%d", lastIndex)
		select {
		case <-server.Context().Done():
			return server.Context().Err()
		case <-c:
			logrus.Debugf("Subscribe.收到notify")
		}
	}
}

func sendSubscribeResponse(index uint64, server protocol.EventService_SubscribeServer, expression *govaluate.EvaluableExpression) error {
	logrus.Debugf("当前正在发送Index:%d", index)
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
	parseUint, err := strconv.ParseUint(meta[2], 10, 64)
	if err != nil {
		return err
	}
	param := make(map[string]interface{})
	entryData := log.Data[1:]

	param["streamName"] = meta[0]
	param["streamId"] = meta[1]
	param["eventId"] = parseUint

	if expression != nil {
		evaluate, err := expression.Evaluate(param)
		if err != nil {
			return err
		}
		b, ok := evaluate.(bool)
		if !ok {
			return fmt.Errorf("表达式错误,返回值不为bool,当前返回值为:%+v", b)
		}

		if !b {
			return nil
		}
	}
	logrus.Debugf("当前正在发送Index:%d", index)
	return server.Send(&protocol.ReadResponse{
		StreamName: meta[0],
		StreamId:   meta[1],
		EventId:    parseUint,
		Data:       entryData,
		Offset:     index,
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
	parseUint, err := strconv.ParseUint(meta[2], 10, 64)
	if err != nil {
		return nil, err
	}
	return &protocol.ReadResponse{
		StreamName: meta[0],
		StreamId:   meta[1],
		EventId:    parseUint,
		Data:       log.Data[1:],
		Offset:     dataIndex,
	}, nil
}

func NewEventService() protocol.EventServiceServer {
	return &EventService{}
}
