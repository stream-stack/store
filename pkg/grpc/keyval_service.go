package grpc

import (
	"context"
	"github.com/dgraph-io/badger/v4"
	v1 "github.com/stream-stack/common/cloudevents.io/genproto/v1"
	"github.com/stream-stack/store/pkg/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newPrivateKeyValueServiceServer() *KeyValueService {
	return &KeyValueService{}
}

type KeyValueService struct {
}

func (k *KeyValueService) Put(ctx context.Context, request *v1.PutRequest) (*v1.PutResponse, error) {
	if err := store.KvStore.Update(func(txn *badger.Txn) error {
		return txn.Set(request.GetKey(), request.GetValue())
	}); err != nil {
		return &v1.PutResponse{Ok: false}, status.Error(codes.InvalidArgument, err.Error())
	}
	return &v1.PutResponse{Ok: true}, nil
}

func (k *KeyValueService) Get(ctx context.Context, request *v1.GetRequest) (*v1.GetResponse, error) {
	response := &v1.GetResponse{}
	if err := store.KvStore.View(func(txn *badger.Txn) error {
		item, err := txn.Get(request.GetKey())
		if err != nil {
			if badger.ErrKeyNotFound == err {
				return nil
			}
			return err
		}
		return item.Value(func(val []byte) error {
			response.Value = val
			return nil
		})
	}); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return response, nil
}
