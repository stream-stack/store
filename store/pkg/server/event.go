package server

import (
	"context"
	"github.com/stream-stack/store/store/common/proto"
	"github.com/stream-stack/store/store/pkg/storage"
	"log"
)

type EventService struct {
}

func (s *EventService) Get(ctx context.Context, request *proto.GetRequest) (*proto.GetResponse, error) {
	data, err := storage.SaveStorage.Get(ctx, request.StreamName, request.StreamId, request.EventId)
	if err != nil {
		return &proto.GetResponse{
			Message: err.Error(),
		}, err
	}
	return &proto.GetResponse{
		EventId: request.EventId,
		Data:    data,
	}, nil
}

func (s *EventService) Save(ctx context.Context, request *proto.SaveRequest) (*proto.SaveResponse, error) {
	log.Printf("[grpc]handler save request{%+v}", request)

	if err := storage.SaveStorage.Save(ctx, request.StreamName, request.StreamId, request.EventId, request.Data); err != nil {
		return &proto.SaveResponse{
			Ack:     false,
			Message: err.Error(),
		}, err
	}
	return &proto.SaveResponse{
		Ack: true,
	}, nil
}
