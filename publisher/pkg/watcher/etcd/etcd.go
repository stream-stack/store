package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/stream-stack/monitor/pkg/watcher"
	"log"
)

const StoreType = "ETCD"
const StorePrefix = "stream"

type etcdConnect struct {
	address string
	cli     *clientv3.Client
}

func (e *etcdConnect) Watch(ctx context.Context) error {
	go func() {
		watch := e.cli.Watch(ctx, "/"+StorePrefix+"/", clientv3.WithPrefix())
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
	var cfg clientv3.Config
	if len(Username) > 0 {
		cfg = clientv3.Config{
			Endpoints:   []string{address},
			DialTimeout: Timeout,
			Context:     ctx,
			Username:    Username,
			Password:    Password,
		}
	} else {
		cfg = clientv3.Config{
			Endpoints:   []string{address},
			DialTimeout: Timeout,
			Context:     ctx,
		}
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

func NewStorageFunc(ctx context.Context, addressSlice []string) ([]watcher.Watcher, error) {
	storages := make([]watcher.Watcher, len(addressSlice))
	for i, address := range addressSlice {
		connect := &etcdConnect{address: address}
		if err := connect.start(ctx, address); err != nil {
			return nil, err
		}
		storages[i] = connect
	}
	return storages, nil
}
