package watcher

import (
	"context"
	"fmt"
	"github.com/stream-stack/publisher/pkg/proto"
	"github.com/stream-stack/publisher/pkg/publisher"
)

type StoreType string

var factoryMap = make(map[StoreType]Factory)

func Register(t StoreType, f Factory) {
	factoryMap[t] = f
}

var WatcherInst Watcher

func Start(ctx context.Context) error {
	//leader 选举
	//连接后端
	factory, ok := factoryMap[StoreType(publisher.BackendTypeValue)]
	if !ok {
		return fmt.Errorf("[storage]not support store type :%v", publisher.BackendTypeValue)
	}
	ss, err := factory(ctx, publisher.BackendAddressValue)
	if err != nil {
		return fmt.Errorf("[storage]init storage error:%v", err)
	}
	WatcherInst = &Proxy{
		backend: ss,
		address: publisher.BackendAddressValue,
	}
	return WatcherInst.Watch(ctx, publisher.In, publisher.WatchStartPoint)
}

type Proxy struct {
	backend []Watcher
	address []string
}

func (p *Proxy) Watch(ctx context.Context, in chan<- proto.SaveRequest, point publisher.StartPoint) error {
	for _, watcher := range p.backend {
		if err := watcher.Watch(ctx, in, point); err != nil {
			return err
		}
	}
	return nil
}
