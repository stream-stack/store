package publisher

import (
	"context"
	"time"
)

type SubscribeEvent interface {
	GetStartPoint() StartPoint
}

type SubscribeManager interface {
	LoadSnapshot(ctx context.Context, streamName string, streamId string) (Snapshot, error)
	Watch(ctx context.Context, point StartPoint, Key string, unmarshaler EventUnmarshaler) (<-chan SubscribeEvent, error)
	SaveSnapshot(ctx context.Context, streamName string, streamId string, data []byte) error
}

type EventUnmarshaler func(key string, data interface{}) ([]SubscribeEvent, error)

type Snapshot interface {
	GetStartPoint() StartPoint
	GetExtData() []byte
	SetStartPoint(i StartPoint)
}

type Subscriber interface {
	GetName() string
	GetUrl() string
	GetStartPoint() StartPoint
	GetKey() string
}

type SubscribePushSetting struct {
	Url         string        `json:"url"`
	ErrIdle     time.Duration `json:"err_idle"`
	MaxErrCount uint8         `json:"max_err_count"`
}

type SubscribeSaveSetting struct {
	Interval time.Duration `json:"interval"`
}

type BaseSubscribe struct {
	Name                 string     `json:"name"`
	StartPoint           StartPoint `json:"start_point"`
	Key                  string     `json:"key"`
	SubscribePushSetting `json:"-"`
	SubscribeSaveSetting `json:"-"`

	ctx              context.Context
	subscribeManager SubscribeManager
}

func (b *BaseSubscribe) GetName() string {
	return b.Name
}

func (b *BaseSubscribe) GetUrl() string {
	return b.Url
}

func (b *BaseSubscribe) GetStartPoint() StartPoint {
	return b.StartPoint
}

func (b *BaseSubscribe) GetKey() string {
	return b.Key
}

type SubscribeManagerFactory func(ctx context.Context, storeAddress []string) (SubscribeManager, error)

type StartPoint interface {
}
