package grpc

import (
	"context"
	v1 "github.com/stream-stack/common/cloudevents.io/genproto/v1"
)

type subscribeHandler interface {
	context() context.Context
	handler(response *v1.CloudEventResponse) error
}

type grpcSubscribeHandler struct {
	server v1.Subscription_SubscribeServer
}

func (g *grpcSubscribeHandler) context() context.Context {
	return g.server.Context()
}

func (g *grpcSubscribeHandler) handler(response *v1.CloudEventResponse) error {
	return g.server.Send(response)
}

type indexSubscribeHandler struct {
	ctx context.Context
}

func (i *indexSubscribeHandler) context() context.Context {
	return i.ctx
}

func (i *indexSubscribeHandler) handler(response *v1.CloudEventResponse) error {
	//TODO:建立索引
	return nil
}
