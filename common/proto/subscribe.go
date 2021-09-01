package proto

import "time"

type SubscribeOperationType string

const Create SubscribeOperationType = "create"
const Delete SubscribeOperationType = "delete"

type SubscribeOperation struct {
	//创建,删除订阅
	Operation SubscribeOperationType `json:"operation"`
	//基础信息
	*BaseSubscribe `json:"base_subscribe"`
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
	SubscribePushSetting `json:"subscribe_push_setting"`
	SubscribeSaveSetting `json:"subscribe_save_setting"`
}

type StartPoint interface {
}
