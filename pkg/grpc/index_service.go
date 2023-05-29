package grpc

import (
	"context"
	"github.com/dgraph-io/badger/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stream-stack/common"
	v1 "github.com/stream-stack/common/cloudevents.io/genproto/v1"
	"github.com/stream-stack/common/util"
	"github.com/stream-stack/store/pkg/store"
	"time"
)

type indexService struct {
	*SubscriptionService
	offset uint64
	ctx    context.Context
}

func (i *indexService) context() context.Context {
	return i.ctx
}

func (i *indexService) handler(response *v1.CloudEventResponse) error {
	key := util.FormatKeyWithEvent(response.Event)
	val := util.FormatKeyWithEventTimestamp(response.Event)
	logrus.Debugf("[grpc][index]index event:%s,val:%s", key, val)
	return store.KvStore.Update(func(txn *badger.Txn) error {
		return txn.Set(key, val)
	})
}

func (i *indexService) Start(ctx context.Context) error {
	i.ctx = ctx
	//read offset from store
	if err := store.KvStore.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("LocalIndexOffset"))
		if err != nil {
			if badger.ErrKeyNotFound == err {
				return nil
			}
			return err
		}
		return item.Value(func(val []byte) error {
			i.offset = util.BytesToUint64(val)
			return nil
		})
	}); err != nil {
		return err
	}

	interval := viper.GetDuration(`OffsetStoreInterval`)
	//start offset save goroutine
	i.saveOffsetWithInterval(ctx, interval)
	//start subscribe store service events
	i.startStoreServiceEventHandler()
	return nil
}

func (i *indexService) saveOffsetWithInterval(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				//save offset to store
				if err := store.KvStore.Update(func(txn *badger.Txn) error {
					return txn.Set([]byte("LocalIndexOffset"), util.Uint64ToBytes(i.offset))
				}); err != nil {
					logrus.Errorf("[grpc][index]save offset to store error: %s", err.Error())
				}
			}
		}
	}()
}

func (i *indexService) startStoreServiceEventHandler() {
	go func() {
		err := i.subscribe(i, i.offset, []byte(common.StoreTypeValuePrefix), nil)
		if err != nil {
			logrus.Errorf("[grpc][index]start subscribe store service events error: %s", err.Error())
		}
	}()
}
