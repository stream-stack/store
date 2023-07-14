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
		op:          make(chan func([]chan interface{})),
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
	logrus.Debugf("[SubscribeWithHandler]begin SubscribeWithHandler, param:{type:%v,eventType:%v,offset:%v}",
		request.GetType(), request.GetEventTypeReg(), offset)
	handler := &grpcSubscribeHandler{server: server}

	return s.SubscribeWithHandler(handler, func(options badger.IteratorOptions) badger.IteratorOptions {
		options.Prefix = prefix
		options.SinceTs = offset + 1
		options.AllVersions = false
		return options
	}, filter)
}

func (s *EventService) Get(ctx context.Context, get *v1.EventGetRequest) (*v1.CloudEventResponse, error) {
	v := &v1.CloudEventStoreRequest{}
	var offset uint64
	key := util.FormatKeyWithEventGetRequest(get)
	if err := store.KvStore.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		offset = item.Version()
		return item.Value(func(val []byte) error {
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
		Offset:    offset,
		Event:     v.Event,
		Key:       key,
		Slot:      v.Slot,
		EventType: v.EventType,
		Timestamp: v.Timestamp,
	}, nil
}

func (s *EventService) ReStore(ctx context.Context, request *v1.CloudEventStoreRequest) (*v1.CloudEventStoreResult, error) {
	key, err := util.FormatKeyWithRequest(request)
	if err != nil {
		return &v1.CloudEventStoreResult{Message: err.Error()}, status.Error(codes.InvalidArgument, err.Error())
	}
	marshal, err := proto.Marshal(request)
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

func (s *EventService) Store(ctx context.Context, request *v1.CloudEventStoreRequest) (*v1.CloudEventStoreResult, error) {
	logicKey, err := util.FormatKeyWithRequest(request)
	if err != nil {
		return &v1.CloudEventStoreResult{Message: err.Error()}, status.Error(codes.InvalidArgument, err.Error())
	}
	marshal, err := proto.Marshal(request)
	if err != nil {
		return &v1.CloudEventStoreResult{Message: err.Error()}, status.Error(codes.InvalidArgument, err.Error())
	}
	var item *badger.Item
	if err = store.KvStore.Update(func(txn *badger.Txn) error {
		if item, err = txn.Get(logicKey); err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
			return txn.Set(logicKey, marshal)
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

func (s *EventService) SubscribeWithHandler(handler subscribeHandler, custom func(options badger.IteratorOptions) badger.IteratorOptions, filter *regexp.Regexp) error {
	options := badger.DefaultIteratorOptions
	if custom != nil {
		options = custom(options)
	}
	current := options.SinceTs
	var err error
	for {
		notify := s.registerSubscriber()
		logrus.Debugf("[SubscribeWithHandler]begin SubscribeWithHandler, param:{offset:%v}", current)
		if current, err = iter(options, handler, filter); err != nil {
			return err
		}

		select {
		case <-handler.Context().Done():
			return nil
		case <-notify:
			logrus.Debugf("[SubscribeWithHandler]notify subscriber,do send data")
		}
	}
}

func getEventTypeReg(request *v1.SubscribeRequest) (*regexp.Regexp, error) {
	if request.EventTypeReg == nil {
		return nil, nil
	}
	return regexp.Compile(request.GetEventTypeReg())
}

func iter(option badger.IteratorOptions, handler subscribeHandler, filter *regexp.Regexp) (uint64, error) {
	var offset uint64
	it := func(txn *badger.Txn) error {
		iterator := txn.NewIterator(option)
		offset = option.SinceTs
		defer iterator.Close()
		for iterator.Rewind(); iterator.Valid(); iterator.Next() {
			item := iterator.Item()
			request := &v1.CloudEventStoreRequest{}
			if err := item.Value(func(val []byte) error {
				if len(val) == 0 {
					return nil
				}
				return proto.Unmarshal(val, request)
			}); err != nil {
				logrus.Errorf("[SubscribeWithHandler]key:%s , unmarshal value error:%v", item.Key(), err.Error())
				return status.Error(codes.Internal, err.Error())
			}
			eventType := request.GetEvent().GetType()
			if filter != nil && !filter.MatchString(eventType) {
				logrus.Debugf("[SubscribeWithHandler]request type %s not match SubscribeWithHandler type regexp, skip", eventType)
				continue
			}
			logrus.Debugf("[SubscribeWithHandler]send cloudevent:{%+v}", request)

			resp := &v1.CloudEventResponse{
				Offset:    item.Version(),
				Event:     request.Event,
				Key:       item.Key(),
				Slot:      request.GetSlot(),
				EventType: request.GetEventType(),
				Timestamp: request.GetTimestamp(),
			}
			if err := handler.Handler(resp); err != nil {
				logrus.Errorf("[SubscribeWithHandler]key:%s,send cloudevent error:%v", item.Key(), err.Error())
				return status.Error(codes.Internal, err.Error())
			} else {
				offset = item.Version()
			}
		}
		return nil
	}
	if err := store.KvStore.View(it); err != nil {
		logrus.Errorf("[SubscribeWithHandler]iterator event error:%v", err)
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
