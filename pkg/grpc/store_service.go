package grpc

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/golang/protobuf/proto"
	"github.com/stream-stack/common"
	v1 "github.com/stream-stack/common/cloudevents.io/genproto/v1"
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
	key := getKey(event)
	if err := store.KvStore.Update(func(txn *badger.Txn) error {
		marshal, err := proto.Marshal(event)
		if err != nil {
			return err
		}
		return txn.Set(key, marshal)
	}); err != nil {
		return &v1.CloudEventStoreResult{Message: err.Error()}, status.Error(codes.InvalidArgument, err.Error())
	}
	s.subService.notify()
	return &v1.CloudEventStoreResult{}, nil
}

func (s *StoreService) Store(ctx context.Context, event *v1.CloudEvent) (*v1.CloudEventStoreResult, error) {
	key := getKey(event)
	if err := store.KvStore.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err == nil {
			return fmt.Errorf("key %s exists", key)
		}
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		//key not found , set key value
		if err == badger.ErrKeyNotFound {
			marshal, err := proto.Marshal(event)
			if err != nil {
				return err
			}
			return txn.Set(key, marshal)
		}
		return nil
	}); err != nil {
		return &v1.CloudEventStoreResult{Message: err.Error()}, status.Error(codes.InvalidArgument, err.Error())
	}
	s.subService.notify()
	return &v1.CloudEventStoreResult{}, nil
}

func getKey(event *v1.CloudEvent) []byte {
	var prefix string
	value, ok := event.GetAttributes()[common.StoreTypeKey]
	if !ok {
		prefix = common.CloudEventStoreTypeValue
	} else {
		prefix = value.GetCeString()
	}

	return []byte(fmt.Sprintf("%s/%s/%s/%s", prefix, event.GetType(), event.GetSource(), event.GetId()))
}
