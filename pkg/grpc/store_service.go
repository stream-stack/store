package grpc

import (
	"context"
	"github.com/dgraph-io/badger/v4"
	"github.com/golang/protobuf/proto"
	v1 "github.com/stream-stack/common/cloudevents.io/genproto/v1"
	"github.com/stream-stack/common/util"
	"github.com/stream-stack/store/pkg/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newStoreService(subService *SubscriptionService) v1.StoreServer {
	return &StoreService{subService: subService}
}

type StoreService struct {
	subService *SubscriptionService
}

func (s *StoreService) ReStore(ctx context.Context, event *v1.CloudEvent) (*v1.CloudEventStoreResult, error) {
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

func (s *StoreService) Store(ctx context.Context, event *v1.CloudEvent) (*v1.CloudEventStoreResult, error) {
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
				equal, err := util.EventEqual(marshal, val)
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
	s.subService.notify()
	return &v1.CloudEventStoreResult{}, nil
}
