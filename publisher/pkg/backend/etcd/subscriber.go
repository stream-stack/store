package etcd

import (
	"context"
	"encoding/json"
	"github.com/stream-stack/publisher/pkg/proto"
	"github.com/stream-stack/publisher/pkg/publisher"
	"google.golang.org/grpc"
	"strconv"
)

func NewSubscribeManagerFunc(ctx context.Context, storeAddress string) (publisher.SubscribeManager, error) {
	dial, err := grpc.Dial(storeAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	go func() {
		select {
		case <-ctx.Done():
			_ = dial.Close()
		}
	}()
	return &SubscribeManagerImpl{grpcClient: proto.NewStorageClient(dial)}, nil
}

type SubscribeManagerImpl struct {
	grpcClient proto.StorageClient
}

func (e *SubscribeManagerImpl) LoadStreamStartPoint(ctx context.Context, streamName string, StreamId string) (publisher.StartPoint, error) {
	get, err := e.grpcClient.Get(ctx, &proto.GetRequest{
		StreamName: streamName,
		StreamId:   StreamId,
		EventId:    publisher.LastEvent,
	})
	if err != nil {
		return nil, err
	}
	m := make(map[string]uint64)
	if err = json.Unmarshal(get.Data, &m); err != nil {
		return nil, err
	}
	return etcdStartPoint(m), nil
}

func (e *SubscribeManagerImpl) LoadAllSubscribe(ctx context.Context) (publisher.SubscriberList, error) {
	get, err := e.grpcClient.Get(context.TODO(), &proto.GetRequest{
		StreamName: publisher.SubscribeStreamName,
		StreamId:   publisher.SubscribeStreamId,
		EventId:    publisher.LastEvent,
	})
	//TODO:如果一个都不存在?
	if err != nil {
		return nil, err
	}
	lastEventId, err := strconv.Atoi(get.EventId)
	if err != nil {
		return nil, err
	}
	dataSlice := make([]*SubscribeEvent, lastEventId+1)
	for i := 0; i < lastEventId; i++ {
		get, err := e.grpcClient.Get(context.TODO(), &proto.GetRequest{
			StreamName: publisher.SubscribeStreamName,
			StreamId:   publisher.SubscribeStreamId,
			EventId:    strconv.Itoa(i),
		})
		if err != nil {
			return nil, err
		}
		s := &SubscribeEvent{}
		if err = json.Unmarshal(get.Data, s); err != nil {
			return nil, err
		}
		dataSlice[i] = s

	}
	s := &SubscribeEvent{}
	if err = json.Unmarshal(get.Data, s); err != nil {
		return nil, err
	}
	dataSlice[lastEventId] = s
	err = rebuildSubscribeMap(ctx, dataSlice, e)
	if err != nil {
		return nil, err
	}

	return map2list(subscribeMap), nil
}

func map2list(m map[string]*publisher.Subscriber) publisher.SubscriberList {
	subscribers := make([]*publisher.Subscriber, len(m))
	i := 0
	for _, subscriber := range m {
		subscribers[i] = subscriber
		i++
	}
	return subscribers
}

var subscribeMap = make(map[string]*publisher.Subscriber)

func rebuildSubscribeMap(ctx context.Context, datas []*SubscribeEvent, e *SubscribeManagerImpl) error {
	subscribeMap = make(map[string]*publisher.Subscriber)
	for _, d := range datas {
		switch d.Operation {
		case SubscribeEventAdd:
			subscribeMap[d.Name] = publisher.NewUrlSubscriber(d.Name, d.Pattern, d.Url, nil)
		case SubscribeEventRemove:
			delete(subscribeMap, d.Name)
		}
	}
	for _, subscriber := range subscribeMap {
		point, err := e.LoadStreamStartPoint(ctx, "subscribe_snapshot", subscriber.Name)
		if err != nil {
			return err
		}
		subscriber.StartPoint = point
	}
	return nil
}
