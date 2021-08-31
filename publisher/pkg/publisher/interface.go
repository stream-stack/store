package publisher

import (
	"context"
	"time"
)

type SubscribeManager interface {
	LoadSnapshot(ctx context.Context, runner *SubscribeRunner) error
	Watch(ctx context.Context, runner *SubscribeRunner) error
	SaveSnapshot(ctx context.Context, runner *SubscribeRunner) error
}

type SubscribeOperation struct {
	//创建,删除订阅
	Operation string `json:"operation"`
	//基础信息
	*BaseSubscribe `json:"-"`
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
}

type SubscribeManagerFactory func(ctx context.Context, storeAddress []string) (SubscribeManager, error)

type StartPoint interface {
}
