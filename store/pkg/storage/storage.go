package storage

import (
	"context"
	"fmt"
	"github.com/stream-stack/store/pkg/config"
	"strings"
)

type StoreType string

var storeFactoryMap = make(map[StoreType]Factory)

func Register(t StoreType, f Factory) {
	storeFactoryMap[t] = f
}

var SaveStorage Storage

func Start(ctx context.Context, c *config.Config) error {
	var p PartitionFunc
	var err error
	switch c.PartitionType {
	case HASHPartitionType:
		p, err = HashPartitionFuncFactory(ctx, c)
		if err != nil {
			return fmt.Errorf("[storage]init PartitionFunc error:%v", err)
		}
	default:
		return fmt.Errorf("[storage]unsupport Partition Type:%v", c.PartitionType)
	}

	//连接后端
	factory, ok := storeFactoryMap[StoreType(c.StoreType)]
	if !ok {
		return fmt.Errorf("[storage]not support store type :%v", c.StoreType)
	}
	ss, err := factory(ctx, c)
	if err != nil {
		return fmt.Errorf("[storage]init storage error:%v", err)
	}
	SaveStorage = &Proxy{
		backend:       ss,
		partitionFunc: p,
		address:       c.StoreAddress,
	}
	return nil
}

type Proxy struct {
	backend       []Storage
	partitionFunc PartitionFunc
	address       []string
}

func (p *Proxy) Get(ctx context.Context, streamName, streamId, eventId string) ([]byte, error) {
	backend, err := p.partitionFunc(streamName, streamId, eventId, p.backend)
	if err != nil {
		return nil, err
	}
	return backend.Get(ctx, streamName, streamId, eventId)
}

func (p *Proxy) GetAddress() string {
	return strings.Join(p.address, ",")
}

func (p *Proxy) Save(ctx context.Context, streamName, streamId, eventId string, data []byte) error {
	backend, err := p.partitionFunc(streamName, streamId, eventId, p.backend)
	if err != nil {
		return err
	}
	return backend.Save(ctx, streamName, streamId, eventId, data)
}
