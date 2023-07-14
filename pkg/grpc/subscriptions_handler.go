package grpc

import (
	"context"
	v1 "github.com/stream-stack/common/cloudevents.io/genproto/v1"
)

type SubscriberFactory func(kvsvc *KeyValueService, evsvc *EventService) Subscriber

type Subscriber interface {
	Start(ctx context.Context) error
	subscribeHandler
}

var subscriberFactories = make([]SubscriberFactory, 0)

func RegisterSubscriberFactory(factory SubscriberFactory) {
	subscriberFactories = append(subscriberFactories, factory)
}

type subscribeHandler interface {
	Context() context.Context
	Handler(response *v1.CloudEventResponse) error
}

type grpcSubscribeHandler struct {
	server v1.PublicEventService_SubscribeServer
}

func (g *grpcSubscribeHandler) Context() context.Context {
	return g.server.Context()
}

func (g *grpcSubscribeHandler) Handler(response *v1.CloudEventResponse) error {
	return g.server.Send(response)
}
