package storage

import (
	"context"
	"fmt"
	"time"
)

type StoreType string

var storeFactoryMap = make(map[StoreType]Factory)

func Register(t StoreType, f Factory) {
	storeFactoryMap[t] = f
}

var SaveStorage Storage

func Start(ctx context.Context) error {
	var err error
	//连接后端
	factory, ok := storeFactoryMap[StoreType(BackendTypeValue)]
	if !ok {
		return fmt.Errorf("[storage]not support store type :%v", BackendTypeValue)
	}
	ss, err := factory(ctx, BackendAddressValue)
	if err != nil {
		return fmt.Errorf("[storage]init storage error:%v", err)
	}
	SaveStorage = ss
	return nil
}

//
//type Proxy struct {
//	backend Storage
//	address []string
//}
//
//func (p *Proxy) Get(ctx context.Context, streamName, streamId, eventId string) ([]byte, error) {
//	return p.backend.Get(ctx, streamName, streamId, eventId)
//}
//
//func (p *Proxy) GetAddress() string {
//	return strings.Join(p.address, ",")
//}
//
//func (p *Proxy) Save(ctx context.Context, streamName, streamId, eventId string, data []byte) error {
//
//	backend, err := p.partitionFunc(streamName, streamId, eventId, p.backend)
//	if err != nil {
//		return err
//	}
//	record := &SaveRecord{
//		Data:       data,
//		CreateTime: time.Now(),
//	}
//	marshal, err := json.Marshal(record)
//	if err != nil {
//		return err
//	}
//	return backend.Save(ctx, streamName, streamId, eventId, marshal)
//}

type SaveRecord struct {
	Data       []byte    `json:"data"`
	CreateTime time.Time `json:"create_time"`
}
