package grpc

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/golang/protobuf/proto"
	v12 "github.com/stream-stack/store/pkg/cloudevents.io/genproto/v1"
	"github.com/stream-stack/store/pkg/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newStoreService(subService *SubscriptionService) v12.StoreServer {
	return &StoreService{subService: subService}
}

type StoreService struct {
	subService *SubscriptionService
}

func (s *StoreService) ReStore(ctx context.Context, event *v12.CloudEvent) (*v12.CloudEventStoreResult, error) {
	key := getKey(event)
	if err := store.KvStore.Update(func(txn *badger.Txn) error {
		marshal, err := proto.Marshal(event)
		if err != nil {
			return err
		}
		return txn.Set(key, marshal)
	}); err != nil {
		return &v12.CloudEventStoreResult{Message: err.Error()}, status.Error(codes.InvalidArgument, err.Error())
	}
	s.subService.notify()
	return &v12.CloudEventStoreResult{}, nil
}

func (s *StoreService) Store(ctx context.Context, event *v12.CloudEvent) (*v12.CloudEventStoreResult, error) {
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
		return &v12.CloudEventStoreResult{Message: err.Error()}, status.Error(codes.InvalidArgument, err.Error())
	}
	s.subService.notify()
	return &v12.CloudEventStoreResult{}, nil
}

func getKey(event *v12.CloudEvent) []byte {
	var prefix string
	value, ok := event.GetAttributes()[StoreTypeKey]
	if !ok {
		prefix = CloudEventStoreTypeValue
	} else {
		prefix = value.GetCeString()
	}

	return []byte(fmt.Sprintf("%s/%s/%s/%s", prefix, event.GetType(), event.GetSource(), event.GetId()))
}
