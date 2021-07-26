package watcher

import (
	"context"
	"fmt"
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
	factory, ok := factoryMap[StoreType(StoreTypeValue)]
	if !ok {
		return fmt.Errorf("[storage]not support store type :%v", StoreTypeValue)
	}
	ss, err := factory(ctx, StoreAddressValue)
	if err != nil {
		return fmt.Errorf("[storage]init storage error:%v", err)
	}
	WatcherInst = &Proxy{
		backend: ss,
		address: StoreAddressValue,
	}
	return WatcherInst.Watch(ctx)
}

type Proxy struct {
	backend []Watcher
	address []string
}

func (p *Proxy) Watch(ctx context.Context) error {
	for _, watcher := range p.backend {
		if err := watcher.Watch(ctx); err != nil {
			return err
		}
	}
	return nil
}
