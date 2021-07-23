package watcher

import (
	"context"
	"fmt"
	"github.com/stream-stack/monitor/pkg/config"
)

type StoreType string

var factoryMap = make(map[StoreType]Factory)

func Register(t StoreType, f Factory) {
	factoryMap[t] = f
}

var WatcherInst Watcher

func Start(ctx context.Context, c *config.Config) error {
	//leader 选举
	//连接后端
	factory, ok := factoryMap[StoreType(c.StoreType)]
	if !ok {
		return fmt.Errorf("[storage]not support store type :%v", c.StoreType)
	}
	ss, err := factory(ctx, c)
	if err != nil {
		return fmt.Errorf("[storage]init storage error:%v", err)
	}
	WatcherInst = &Proxy{
		backend: ss,
		address: c.StoreAddress,
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
