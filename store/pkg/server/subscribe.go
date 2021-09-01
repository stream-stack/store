package server

import (
	"context"
	"encoding/json"
	"github.com/stream-stack/store/store/common/proto"
	"github.com/stream-stack/store/store/pkg/storage"
	"time"
)

type SubscribeService struct {
	eventService *EventService
}

func (s *SubscribeService) Create(ctx context.Context, request *proto.CreateRequest) (*proto.CreateResponse, error) {
	//system/subscribe
	duration, err := time.ParseDuration(request.PushMaxRequestDuration)
	if err != nil {
		return &proto.CreateResponse{
			Ack:     false,
			Message: "Parse MaxRequest Duration error:" + err.Error(),
		}, err
	}
	saveInterval, err := time.ParseDuration(request.SaveInterval)
	if err != nil {
		return &proto.CreateResponse{
			Ack:     false,
			Message: "Parse SaveInterval Duration error:" + err.Error(),
		}, err
	}
	var point interface{}
	switch request.StartPoint {
	case proto.CreateRequest_Current:
		point, err = storage.SaveStorage.GetCurrentStartPoint(ctx)
		if err != nil {
			return &proto.CreateResponse{
				Ack:     false,
				Message: "get StartPoint error:" + err.Error(),
			}, err
		}
	case proto.CreateRequest_Begin:
		point = 0
	}

	operation := proto.SubscribeOperation{
		Operation: proto.Create,
		BaseSubscribe: &proto.BaseSubscribe{
			Name:       request.Name,
			StartPoint: point,
			Key:        request.WatchKey,
			SubscribePushSetting: proto.SubscribePushSetting{
				Url:                  request.PushUrl,
				MaxRequestDuration:   duration,
				MaxRequestRetryCount: int(request.PushMaxRequestRetryCount),
			},
			SubscribeSaveSetting: proto.SubscribeSaveSetting{
				Interval: saveInterval,
			},
		},
	}
	marshal, err := json.Marshal(operation)
	if err != nil {
		return &proto.CreateResponse{
			Ack:     false,
			Message: "Marshal request error:" + err.Error(),
		}, err
	}
	_, err = s.eventService.Save(ctx, &proto.SaveRequest{
		StreamName: "system",
		StreamId:   "subscribe",
		EventId:    time.Now().String(),
		Data:       marshal,
	})
	if err != nil {
		return &proto.CreateResponse{
			Ack:     false,
			Message: "save request error:" + err.Error(),
		}, err
	}
	return &proto.CreateResponse{
		Ack: true,
	}, err
}

func (s *SubscribeService) Delete(ctx context.Context, request *proto.DeleteRequest) (*proto.DeleteResponse, error) {
	operation := proto.SubscribeOperation{
		Operation: proto.Delete,
		BaseSubscribe: &proto.BaseSubscribe{
			Name: request.Name,
		},
	}
	marshal, err := json.Marshal(operation)
	if err != nil {
		return &proto.DeleteResponse{
			Ack:     false,
			Message: "Marshal request error:" + err.Error(),
		}, err
	}
	_, err = s.eventService.Save(ctx, &proto.SaveRequest{
		StreamName: "system",
		StreamId:   "subscribe",
		EventId:    time.Now().String(),
		Data:       marshal,
	})
	if err != nil {
		return &proto.DeleteResponse{
			Ack:     false,
			Message: "save request error:" + err.Error(),
		}, err
	}
	return &proto.DeleteResponse{
		Ack: true,
	}, err
}
