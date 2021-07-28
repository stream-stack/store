package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type StoreType string

var storeFactoryMap = make(map[StoreType]Factory)

func Register(t StoreType, f Factory) {
	storeFactoryMap[t] = f
}

var SaveStorage Storage

func Start(ctx context.Context) error {
	var p PartitionFunc
	var err error
	switch PartitionValue {
	case HASHPartitionType:
		p, err = HashPartitionFuncFactory(ctx, BackendAddressValue)
		if err != nil {
			return fmt.Errorf("[storage]init PartitionFunc error:%v", err)
		}
	default:
		return fmt.Errorf("[storage]unsupport Partition Type:%v", PartitionValue)
	}

	//连接后端
	factory, ok := storeFactoryMap[StoreType(BackendTypeValue)]
	if !ok {
		return fmt.Errorf("[storage]not support store type :%v", BackendTypeValue)
	}
	ss, err := factory(ctx, BackendAddressValue)
	if err != nil {
		return fmt.Errorf("[storage]init storage error:%v", err)
	}
	SaveStorage = &Proxy{
		backend:       ss,
		partitionFunc: p,
		address:       BackendAddressValue,
	}
	return nil
}

type Proxy struct {
	backend       []Storage
	partitionFunc PartitionFunc
	address       []string
}

func (p *Proxy) Get(ctx context.Context, streamName, streamId, eventId string) ([]byte, error) {
	if eventId == FirstEvent || eventId == LastEvent {
		m := make(map[time.Time][]byte)
		for _, storage := range p.backend {
			data, err := storage.Get(ctx, streamName, streamId, eventId)
			if err != nil {
				return nil, err
			}
			d := &SaveRecord{}
			if err := json.Unmarshal(data, d); err != nil {
				return nil, err
			}
			m[d.CreateTime] = d.Data
		}
		var current time.Time

		for t := range m {
			//排序,获取最小,最大的time
			if current.IsZero() {
				current = t
			}
			if current.Before(t) && eventId == LastEvent {
				current = t
			}
			if current.After(t) && eventId == FirstEvent {
				current = t
			}
		}
		return m[current], nil
	}

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
	record := &SaveRecord{
		Data:       data,
		CreateTime: time.Now(),
	}
	marshal, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return backend.Save(ctx, streamName, streamId, eventId, marshal)
}

type SaveRecord struct {
	Data       []byte    `json:"data"`
	CreateTime time.Time `json:"create_time"`
}
