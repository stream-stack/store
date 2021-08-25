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

type SaveRecord struct {
	Data       []byte    `json:"data"`
	CreateTime time.Time `json:"create_time"`
}
