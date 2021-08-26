package publisher

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"time"
)

type SubscribeEvent interface {
	GetStartPoint() StartPoint
	GetCloudEvent() cloudevents.Event
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
	Url                  string        `json:"url"`
	MaxRequestDuration   time.Duration `json:"max_request_duration"`
	MaxRequestRetryCount int           `json:"max_request_retry_count"`
}

type SubscribeSaveSetting struct {
	Interval time.Duration `json:"interval"`
}

type BaseSubscribe struct {
	Name                 string     `json:"name"`
	StartPoint           StartPoint `json:"start_point"`
	Key                  string     `json:"watchKey"`
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
