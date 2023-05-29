package grpc

import (
	"bytes"
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
	key := util.FormatKeyWithEventTimestamp(event)
	marshal, err := proto.Marshal(event)
	if err != nil {
		return &v1.CloudEventStoreResult{Message: err.Error()}, status.Error(codes.InvalidArgument, err.Error())
	}
	if err = store.KvStore.Update(func(txn *badger.Txn) error {
		if item, err := txn.Get(key); err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
			return txn.Set(key, marshal)
		} else {
			return item.Value(func(val []byte) error {
				if !bytes.Equal(marshal, val) {
					return util.EventExists
				}
				return nil
			})
		}
	}); err != nil {
		return &v1.CloudEventStoreResult{Message: err.Error()}, status.Error(codes.InvalidArgument, err.Error())
	}
	s.subService.notify()
	return &v1.CloudEventStoreResult{}, nil
}
