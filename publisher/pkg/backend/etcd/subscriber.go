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

const SubscribeEventAdd = "ADD"
const SubscribeEventRemove = "REMOVE"

type SubscribeEvent struct {
	Operation string `json:"operation"`
	Name      string `json:"name"`
	Pattern   string `json:"pattern"`
	Url       string `json:"url"`
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
	rebuildSubscribeMap(dataSlice)

	//map 转list

	return subscribeMap, nil
}

var subscribeMap = make(map[string]*publisher.Subscriber)

func rebuildSubscribeMap(datas []*SubscribeEvent) {
	subscribeMap = make(map[string]*publisher.Subscriber)
	for _, d := range datas {
		switch d.Operation {
		case SubscribeEventAdd:
			subscribeMap[d.Name] = NewSubscriber(d)
		case SubscribeEventRemove:
			delete(subscribeMap, d.Name)
		}
	}
}

func NewSubscriber(d *SubscribeEvent) *publisher.Subscriber {
	return &publisher.Subscriber{
		Name:    d.Name,
		Pattern: d.Pattern,
		Url:     d.Url,
		//TODO:获取startPoint
		//StartPoint: ,
	}
}
