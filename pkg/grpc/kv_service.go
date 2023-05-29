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

func newKVService() v1.KVServer {
	return &KVService{}
}

type KVService struct {
}

func (K *KVService) Get(ctx context.Context, get *v1.KVGet) (*v1.CloudEventResponse, error) {
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
