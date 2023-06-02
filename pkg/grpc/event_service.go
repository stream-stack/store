package grpc

import (
	"context"
	"github.com/dgraph-io/badger/v4"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/stream-stack/common"
	v1 "github.com/stream-stack/common/cloudevents.io/genproto/v1"
	"github.com/stream-stack/common/util"
	"github.com/stream-stack/store/pkg/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"regexp"
)

func newPublicEventServiceServer() *EventService {
	return &EventService{
		subscribers: make([]chan interface{}, 0),
		op:          make(chan func([]chan interface{}), 0),
	}
}

type EventService struct {
	subscribers []chan interface{}
	op          chan func([]chan interface{})
}

func (s *EventService) Subscribe(request *v1.SubscribeRequest, server v1.PublicEventService_SubscribeServer) error {
	var offset uint64
	if request.Offset != nil {
		offset = request.GetOffset()
	}
	prefix := getSubscribePrefix(request)
	filter, err := getEventTypeReg(request)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	logrus.Debugf("[subscribe]begin subscribe, param:{type:%v,eventType:%v,offset:%v}",
		request.GetType(), request.GetEventTypeReg(), offset)
	return s.subscribe(&grpcSubscribeHandler{server: server}, offset, prefix, filter)
}

func (s *EventService) Get(ctx context.Context, get *v1.EventGetRequest) (*v1.CloudEventResponse, error) {
	v := &v1.CloudEvent{}
	var offset uint64
	key := util.FormatKeyWithGet(get)
	if err := store.KvStore.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		var originKey []byte
		if err := item.Value(func(val []byte) error {
			originKey = val
			return nil
		}); err != nil {
			return err
		}
		i, err := txn.Get(originKey)
		if err != nil {
			return err
		}
		offset = i.Version()
		return i.Value(func(val []byte) error {
			return proto.Unmarshal(val, v)
		})
	}); err != nil {
		errCode := codes.Internal
		if err == badger.ErrKeyNotFound {
			errCode = codes.NotFound
		}
		return nil, status.Error(errCode, err.Error())
	}
	return &v1.CloudEventResponse{
		Offset: offset,
		Event:  v,
	}, nil
}

func (s *EventService) ReStore(ctx context.Context, event *v1.CloudEvent) (*v1.CloudEventStoreResult, error) {
	key := util.FormatKeyWithEventTimestamp(event)
	marshal, err := proto.Marshal(event)
	if err != nil {
		return &v1.CloudEventStoreResult{Message: err.Error()}, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := store.KvStore.Update(func(txn *badger.Txn) error {
		return txn.Set(key, marshal)
	}); err != nil {
		return &v1.CloudEventStoreResult{Message: err.Error()}, status.Error(codes.InvalidArgument, err.Error())
	}
	return &v1.CloudEventStoreResult{}, nil
}

func (s *EventService) Store(ctx context.Context, event *v1.CloudEvent) (*v1.CloudEventStoreResult, error) {
	logicKey, err := util.FormatKeyWithEvent(event)
	if err != nil {
		return &v1.CloudEventStoreResult{Message: err.Error()}, status.Error(codes.InvalidArgument, err.Error())
	}
	key := util.FormatKeyWithEventTimestamp(event)
	marshal, err := proto.Marshal(event)
	if err != nil {
		return &v1.CloudEventStoreResult{Message: err.Error()}, status.Error(codes.InvalidArgument, err.Error())
	}
	var item *badger.Item
	if err = store.KvStore.Update(func(txn *badger.Txn) error {
		if item, err = txn.Get(logicKey); err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
			return txn.Set(key, marshal)
		} else {
			var logicItem *badger.Item
			if err = item.Value(func(val []byte) error {
				logicItem, err = txn.Get(val)
				if err != nil {
					return err
				}
				return nil
			}); err != nil {
				return err
			}
			return logicItem.Value(func(val []byte) error {
				equal, err := util.EventEqualWithBytes(marshal, val)
				if err != nil {
					return err
				}
				if equal {
					return nil
				}
				return util.EventExists
			})
		}
	}); err != nil {
		return &v1.CloudEventStoreResult{Message: err.Error()}, status.Error(codes.InvalidArgument, err.Error())
	}
	s.notify()
	return &v1.CloudEventStoreResult{}, nil
}

func (s *EventService) subscribe(handler subscribeHandler, offset uint64, prefix []byte, filter *regexp.Regexp) error {
	current := offset
	var err error
	for {
		notify := s.registerSubscriber()
		logrus.Debugf("[subscribe]begin subscribe, param:{offset:%v}", current)
		if current, err = iter(prefix, current, handler, filter); err != nil {
			return err
		}

		select {
		case <-handler.context().Done():
			return nil
		case <-notify:
			logrus.Debugf("[subscribe]notify subscriber,do send data")
		}
	}
}

func getEventTypeReg(request *v1.SubscribeRequest) (*regexp.Regexp, error) {
	if request.EventTypeReg == nil {
		return nil, nil
	}
	return regexp.Compile(request.GetEventTypeReg())
}

func iter(prefix []byte, offset uint64, handler subscribeHandler, filter *regexp.Regexp) (uint64, error) {
	it := func(txn *badger.Txn) error {
		options := badger.DefaultIteratorOptions
		options.SinceTs = offset + 1
		if prefix != nil && len(prefix) > 0 {
			options.Prefix = prefix
		}
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
			if filter != nil && !filter.MatchString(event.GetType()) {
				logrus.Debugf("[subscribe]event type %s not match subscribe type regexp, skip", event.GetType())
				continue
			}
			logrus.Debugf("[subscribe]send cloudevent:{%+v}", event)

			resp := &v1.CloudEventResponse{
				Offset: item.Version(),
				Event:  event,
			}
			if err := handler.handler(resp); err != nil {
				logrus.Errorf("[subscribe]key:%s,send cloudevent error:%v", item.Key(), err.Error())
				return status.Error(codes.Internal, err.Error())
			} else {
				offset = item.Version()
			}
		}
		return nil
	}
	if err := store.KvStore.View(it); err != nil {
		logrus.Errorf("[subscribe]iterator event error:%v", err)
		return offset, status.Error(codes.Internal, err.Error())
	}
	return offset, nil
}

func getSubscribePrefix(request *v1.SubscribeRequest) []byte {
	tp := common.CloudEventStoreTypeValue
	if request.Type != nil {
		tp = request.GetType()
	}
	return []byte(tp)
}

func (s *EventService) startSubscribeServer(ctx context.Context) {
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

func (s *EventService) registerSubscriber() chan interface{} {
	notify := make(chan interface{})
	s.op <- func(i []chan interface{}) {
		s.subscribers = append(i, notify)
	}
	return notify
}

func (s *EventService) notify() {
	s.op <- func(i []chan interface{}) {
		for _, c := range i {
			close(c)
		}
		s.subscribers = make([]chan interface{}, 0)
	}
}

func (s *EventService) startIndexServer(ctx context.Context, kvsvc *KeyValueService) error {
	is := &indexService{
		evsvc: s,
		kvsvc: kvsvc,
	}
	return is.Start(ctx)
}
