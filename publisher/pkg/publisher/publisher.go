package publisher

import (
	"context"
	"fmt"
	"time"
)

type BackendType string

var factoryMap = make(map[BackendType]SubscribeManagerFactory)

func Register(t BackendType, f SubscribeManagerFactory) {
	factoryMap[t] = f
}

var eventUnmarshalMap = make(map[BackendType]EventUnmarshaler)

func RegisterEventUnmarshal(t BackendType, f EventUnmarshaler) {
	eventUnmarshalMap[t] = f
}

var dataUnmarshalMap = make(map[BackendType]DataUnmarshaler)

func RegisterExtDataUnmarshal(t BackendType, f DataUnmarshaler) {
	dataUnmarshalMap[t] = f
}

type action func(ctx context.Context, event SubscribeEvent, data interface{}, pushSetting SubscribePushSetting) error

type DataUnmarshaler func(data []byte) (interface{}, error)

var ss SubscribeManager
var eu EventUnmarshaler
var du DataUnmarshaler

func Start(ctx context.Context) error {
	var err error
	factory, ok := factoryMap[BackendType(BackendTypeValue)]
	if !ok {
		return fmt.Errorf("[publisher]not support store type :%v", BackendTypeValue)
	}
	eu, ok = eventUnmarshalMap[BackendType(BackendTypeValue)]
	if !ok {
		return fmt.Errorf("[publisher]not support eventUnmarshaler store type :%v", BackendTypeValue)
	}
	du, ok = dataUnmarshalMap[BackendType(BackendTypeValue)]
	if !ok {
		return fmt.Errorf("[publisher]not support DataUnmarshaler store type :%v", BackendTypeValue)
	}

	ss, err = factory(ctx, BackendAddressValue)
	if err != nil {
		return fmt.Errorf("[publisher]init storage error:%v", err)
	}
	runner := NewSubscribeRunner(ctx, ss, systemDataUnmarshaler, eu, "system", "snapshot",
		"system_subscribe", "system/subscribe", systemAction, time.Minute)
	return runner.start()
}
