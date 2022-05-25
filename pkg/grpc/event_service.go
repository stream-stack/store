package grpc

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Knetic/govaluate"
	hraft "github.com/hashicorp/raft"
	"github.com/sirupsen/logrus"
	protocol "github.com/stream-stack/common/protocol/store"
	"github.com/stream-stack/store/pkg/raft"
	"strconv"
)

type EventService struct {
}

func (e *EventService) Subscribe(request *protocol.SubscribeRequest, server protocol.EventService_SubscribeServer) error {
	logrus.Debugf(`收到订阅请求%+v`, request)
	var expression *govaluate.EvaluableExpression
	var err error
	if len(request.Regexp) != 0 {
		expression, err = govaluate.NewEvaluableExpression(request.Regexp)
		if err != nil {
			return err
		}
	}
	var start = request.Offset
	id := request.SubscribeId
	if start == 0 {
		logrus.Debugf("subscribe:%v,set start to 1", id)
		start = 1
	}
	filter := func(log *hraft.Log) (bool, error) {
		if expression == nil {
			return true, nil
		}
		meta, err := protocol.ParseMeta(log.Extensions)
		if err != nil {
			logrus.Debugf("key %s parse meta error:%v,filter skip", log.Extensions, err)
			return false, nil
		}
		eventId, err := strconv.ParseUint(meta[2], 10, 64)
		if err != nil {
			return false, err
		}
		param := make(map[string]interface{})
		param["streamName"] = meta[0]
		param["streamId"] = meta[1]
		param["eventId"] = eventId
		logrus.Debugf(`当前事件streamName:%v,streamId:%v,eventId:%d`, meta[0], meta[1], eventId)

		evaluate, err := expression.Evaluate(param)
		logrus.Debugf(`表达式 %s 执行结果:%v`, expression.String(), evaluate)
		if err != nil {
			return false, err
		}
		b, ok := evaluate.(bool)
		if !ok {
			return false, fmt.Errorf("表达式错误,返回值不为bool,当前返回值为:%+v", b)
		}

		return b, nil
	}
	handler := func(key, value []byte, version uint64) error {
		start = version
		meta, err := protocol.ParseMeta(key)
		if err != nil {
			logrus.Errorf("key %s parse meta error:%v,handler skip", key, err)
			return nil
		}
		parseUint, err := strconv.ParseUint(meta[2], 10, 64)
		if err != nil {
			return err
		}
		param := make(map[string]interface{})
		param["streamName"] = meta[0]
		param["streamId"] = meta[1]
		param["eventId"] = parseUint

		return server.Send(&protocol.ReadResponse{
			StreamName: meta[0],
			StreamId:   meta[1],
			EventId:    parseUint,
			Data:       value,
			Offset:     version,
		})
	}
	for {
		logrus.Debugf("subscribe:%v,offset:%d", id, start)
		//注册channel
		notify := raft.Wait(filter, id)
		logrus.Debugf("subscribe:%v,registed notify channel", id)

		if err := raft.Iterator([]byte(""), start, filter, handler); err != nil {
			logrus.Errorf("subscribe:%v,iterator error:%v", id, err)
			return err
		}
		logrus.Debugf("subscribe:%v,iterator done,wait new event", id)

		select {
		case <-server.Context().Done():
			return server.Context().Err()
		case <-notify:
			logrus.Debugf("subscribe:%v,notify channel receive", id)
		}
	}
}

func (e *EventService) Apply(ctx context.Context, request *protocol.ApplyRequest) (*protocol.ApplyResponse, error) {
	logrus.Debugf(`apply StreamName:%v,StreamId:%v,EventId:%v`, request.StreamName, request.StreamId, request.EventId)
	key := protocol.FormatApplyMeta(request.StreamName, request.StreamId, request.EventId)
	data, offset, err := raft.GetLogByKey(key)

	if err != nil && raft.ErrKeyNotFound == err {
		applyFuture := raft.Raft.ApplyLog(hraft.Log{
			Data:       request.Data,
			Extensions: key,
		}, applyLogTimeout)
		if err := applyFuture.Error(); err != nil {
			return &protocol.ApplyResponse{
				Ack:     true,
				Message: err.Error(),
			}, err
		}
		return &protocol.ApplyResponse{
			Ack: true,
		}, nil
	}

	if err != nil {
		return &protocol.ApplyResponse{
			Ack:     false,
			Message: err.Error(),
		}, err
	}
	//判断data与request.Data是否相等
	if bytes.Equal(data, request.Data) {
		return &protocol.ApplyResponse{
			Ack:     true,
			Message: strconv.FormatUint(offset, 10),
		}, nil
	}
	err = fmt.Errorf("event exist,offset is %d", offset)
	return &protocol.ApplyResponse{
		Ack:     false,
		Message: err.Error(),
	}, err
}

func (e *EventService) Read(ctx context.Context, request *protocol.ReadRequest) (*protocol.ReadResponse, error) {
	key := protocol.FormatApplyMeta(request.StreamName, request.StreamId, request.EventId)
	logrus.Debugf(`read StreamName:%v,StreamId:%v,EventId:%v`, request.StreamName, request.StreamId, request.EventId)
	data, offset, err := raft.GetLogByKey(key)
	if err != nil {
		return nil, err
	}
	return &protocol.ReadResponse{
		StreamName: request.StreamName,
		StreamId:   request.StreamId,
		EventId:    request.EventId,
		Data:       data,
		Offset:     offset,
	}, nil
}

func NewEventService() protocol.EventServiceServer {
	return &EventService{}
}
