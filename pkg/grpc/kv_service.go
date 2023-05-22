package grpc

import (
	"context"
	"github.com/dgraph-io/badger/v4"
	"github.com/golang/protobuf/proto"
	v12 "github.com/stream-stack/store/pkg/cloudevents.io/genproto/v1"
	"github.com/stream-stack/store/pkg/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newKVService() v12.KVServer {
	return &KVService{}
}

type KVService struct {
}

func (K *KVService) Get(ctx context.Context, get *v12.KVGet) (*v12.CloudEventResponse, error) {
	v := &v12.CloudEvent{}
	var offset uint64
	if err := store.KvStore.View(func(txn *badger.Txn) error {
		item, err := txn.Get(get.GetKey())
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
	return &v12.CloudEventResponse{
		Offset: offset,
		Event:  v,
	}, nil
}
