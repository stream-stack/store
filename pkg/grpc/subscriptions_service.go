package grpc

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/stream-stack/common"
	v1 "github.com/stream-stack/common/cloudevents.io/genproto/v1"
	"github.com/stream-stack/store/pkg/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newSubscriptionsService() *SubscriptionService {
	return &SubscriptionService{
		subscribers: make([]chan interface{}, 0),
		op:          make(chan func([]chan interface{}), 0),
	}
}

type SubscriptionService struct {
	subscribers []chan interface{}
	op          chan func([]chan interface{})
}

func (s *SubscriptionService) Subscribe(request *v1.SubscribeRequest, server v1.Subscription_SubscribeServer) error {
	var offset uint64
	if request.Offset != nil {
		offset = request.GetOffset()
	}
	prefix := getSubscribePrefix(request)
	for {
		notify := s.registerSubscriber()
		logrus.Debugf("[subscribe]register subscriber{type:%v,eventType:%v,offset:%v} success",
			request.GetType(), request.GetEventType(), offset)

		if err := iter(prefix, offset, server); err != nil {
			return err
		}

		select {
		case <-server.Context().Done():
			return nil
		case <-notify:
			logrus.Debugf("[subscribe]notify subscriber,do send data")
		}
	}
}

func iter(prefix []byte, offset uint64, server v1.Subscription_SubscribeServer) error {
	handler := func(txn *badger.Txn) error {
		options := badger.DefaultIteratorOptions
		options.SinceTs = offset
		options.Prefix = prefix
		iterator := txn.NewIterator(options)
		defer iterator.Close()
		for iterator.Rewind(); iterator.Valid(); iterator.Next() {
			item := iterator.Item()
			event := &v1.CloudEvent{}
			if err := item.Value(func(val []byte) error {
				return proto.Unmarshal(val, event)
			}); err != nil {
				logrus.Errorf("[subscribe]key:%s , unmarshal value error:%v", item.Key(), err.Error())
				return status.Error(codes.Internal, err.Error())
			}
			if err := server.Send(&v1.CloudEventResponse{
				Offset: item.Version(),
				Event:  event,
			}); err != nil {
				logrus.Errorf("[subscribe]key:%s,send cloudevent error:%v", item.Key(), err.Error())
				return status.Error(codes.Internal, err.Error())
			}
		}
		return nil
	}
	if err := store.KvStore.View(handler); err != nil {
		logrus.Errorf("[subscribe]iterator event error:%v", err)
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func getSubscribePrefix(request *v1.SubscribeRequest) []byte {
	tp := common.CloudEventStoreTypeValue
	if request.Type == nil {
		tp = request.GetType()
	}
	if request.EventType == nil {
		return []byte(tp)
	}
	return []byte(fmt.Sprintf(`%s/%s`, tp, request.GetEventType()))
}

func (s *SubscriptionService) start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case f := <-s.op:
				f(s.subscribers)
			}
		}
	}()
}

func (s *SubscriptionService) registerSubscriber() chan interface{} {
	notify := make(chan interface{})
	s.op <- func(i []chan interface{}) {
		s.subscribers = append(i, notify)
	}
	return notify
}

func (s *SubscriptionService) notify() {
	s.op <- func(i []chan interface{}) {
		for _, c := range i {
			close(c)
		}
		s.subscribers = make([]chan interface{}, 0)
	}
}
