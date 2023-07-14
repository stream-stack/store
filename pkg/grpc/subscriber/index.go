package subscriber

import (
	"context"
	"github.com/dgraph-io/badger/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stream-stack/common"
	v1 "github.com/stream-stack/common/cloudevents.io/genproto/v1"
	"github.com/stream-stack/common/util"
	"github.com/stream-stack/store/pkg/grpc"
	"github.com/stream-stack/store/pkg/store"
	"time"
)

const LocalIndexOffsetKey = "LocalIndexOffset"

type indexSubscriber struct {
	evsvc  *grpc.EventService
	kvsvc  *grpc.KeyValueService
	offset uint64
	ctx    context.Context
}

func newIndexSubscriber(evsvc *grpc.EventService, kvsvc *grpc.KeyValueService) *indexSubscriber {
	return &indexSubscriber{
		evsvc: evsvc,
		kvsvc: kvsvc,
	}
}

func (i *indexSubscriber) Context() context.Context {
	return i.ctx
}

func (i *indexSubscriber) Handler(response *v1.CloudEventResponse) error {
	//TODO: 批量写入
	key, err := util.FormatKeyWithEvent(response.Event)
	if err != nil {
		return err
	}
	val := response.GetKey()
	logrus.Debugf("[grpc][index]index event:%s,val:%s", key, val)
	i.offset = response.Offset
	return store.KvStore.Update(func(txn *badger.Txn) error {
		return txn.Set(key, val)
	})
}

func (i *indexSubscriber) Start(ctx context.Context) error {
	i.ctx = ctx
	get, err := i.kvsvc.Get(ctx, &v1.GetRequest{Key: []byte(LocalIndexOffsetKey)})
	if err != nil {
		return err
	}
	if get.Value != nil {
		i.offset = util.BytesToUint64(get.Value)
	}

	interval := viper.GetDuration(`IndexOffsetStoreInterval`)
	//start offset save goroutine
	i.saveOffsetWithInterval(ctx, interval)
	//start subscribe store service events
	i.startStoreServiceEventHandler()
	return nil
}

func (i *indexSubscriber) saveOffsetWithInterval(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	prevOffset := i.offset
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if prevOffset == i.offset {
					logrus.Debugf("[grpc][index]offset not change,skip save offset to store")
					continue
				}
				//save offset to store
				_, err := i.kvsvc.Put(ctx, &v1.PutRequest{
					Key:   []byte(LocalIndexOffsetKey),
					Value: util.Uint64ToBytes(i.offset),
				})
				if err != nil {
					logrus.Errorf("[grpc][index]save offset to store error: %s", err.Error())
				} else {
					logrus.Debugf("[grpc][index]save offset to store success,offset:%d", i.offset)
				}
				prevOffset = i.offset
			}
		}
	}()
}

func (i *indexSubscriber) startStoreServiceEventHandler() {
	go func() {
		err := i.evsvc.SubscribeWithHandler(i, func(options badger.IteratorOptions) badger.IteratorOptions {
			//i.offset, []byte(common.StoreTypeValuePrefix)
			options.SinceTs = i.offset + 1
			options.Prefix = []byte(common.StoreTypeValuePrefix)
			return options
		}, nil)
		if err != nil {
			logrus.Errorf("[grpc][index]start subscribe store service events error: %s", err.Error())
		}
	}()
}
