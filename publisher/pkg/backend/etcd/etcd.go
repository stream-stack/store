package etcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/stream-stack/publisher/pkg/proto"
	"github.com/stream-stack/publisher/pkg/publisher"
	"github.com/stream-stack/publisher/pkg/watcher"
	"log"
)

const BackendType = "ETCD"
const BackendPrefix = "stream"

type etcdConnect struct {
	address string
	cli     *clientv3.Client
}

func (e *etcdConnect) Watch(ctx context.Context, in chan<- proto.SaveRequest, point publisher.StartPoint) error {
	go func() {
		//TODO:使用startpoint
		watch := e.cli.Watch(ctx, "/"+BackendPrefix+"/", clientv3.WithPrefix(), clientv3.WithRev(0))
		for {
			select {
			case <-ctx.Done():
				return
			case response := <-watch:
				for _, event := range response.Events {
					log.Printf("[watcher]收到事件,%v,%v,%v", event.Type, string(event.Kv.Key), string(event.Kv.Value))
					var name, sid, id string
					_, err := fmt.Sscanf(string(event.Kv.Key), "/"+BackendPrefix+"/%s/%s/%s", &name, &sid, &id)
					if err != nil {
						log.Printf("[watcher]解析key:%s,错误:%v", string(event.Kv.Key), err)
						continue
					}
					in <- proto.SaveRequest{
						StreamName: name,
						StreamId:   sid,
						EventId:    id,
						Data:       event.Kv.Value,
					}
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
