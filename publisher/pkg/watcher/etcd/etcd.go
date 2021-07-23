package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/stream-stack/monitor/pkg/config"
	"github.com/stream-stack/monitor/pkg/watcher"
	"log"
	"time"
)

const StoreType = "ETCD"

func init() {
	watcher.Register(StoreType, NewStorageFunc)
}

type etcdConnect struct {
	address string
	cli     *clientv3.Client
}

func (e *etcdConnect) Watch(ctx context.Context) error {
	go func() {
		watch := e.cli.Watch(ctx, "/", clientv3.WithPrefix())
		for {
			select {
			case <-ctx.Done():
				return
			case response := <-watch:
				for _, event := range response.Events {
					log.Printf("[watcher]收到事件,%v,%v,%v", event.Type, string(event.Kv.Key), string(event.Kv.Value))
				}
			}
		}

	}()
	return nil
}

func (e *etcdConnect) GetAddress() string {
	return e.address
}

func (e *etcdConnect) start(ctx context.Context, address string) error {
	var err error
	cfg := clientv3.Config{
		Endpoints: []string{address},
		//TODO:超时设置
		DialTimeout: 5 * time.Second,
		Context:     ctx,
		Username:    "root",
		Password:    "t1aZnKJLoR",
	}
	e.cli, err = clientv3.New(cfg)
	if err != nil {
		return err
	}
	go func() {
		select {
		case <-ctx.Done():
			_ = e.cli.Close()
		}
	}()
	return nil
}

func NewStorageFunc(ctx context.Context, c *config.Config) ([]watcher.Watcher, error) {
	storages := make([]watcher.Watcher, len(c.StoreAddress))
	for i, address := range c.StoreAddress {
		connect := &etcdConnect{address: address}
		if err := connect.start(ctx, address); err != nil {
			return nil, err
		}
		storages[i] = connect
	}
	return storages, nil
}
