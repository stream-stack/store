package publisher

import (
	"context"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"time"
)

type BackendType string

var factoryMap = make(map[BackendType]SubscribeManagerFactory)

func Register(t BackendType, f SubscribeManagerFactory) {
	factoryMap[t] = f
}

type Action func(ctx context.Context, event cloudevents.Event, runner *SubscribeRunner) error

var ss SubscribeManager

func Start(ctx context.Context) error {
	var err error
	factory, ok := factoryMap[BackendType(BackendTypeValue)]
	if !ok {
		return fmt.Errorf("[publisher]not support store type :%v", BackendTypeValue)
	}

	ss, err = factory(ctx, BackendAddressValue)
	if err != nil {
		return fmt.Errorf("[publisher]init storage error:%v", err)
	}
	runner := newSystemSubscribeRunner(ctx, ss, "system", "snapshot", "system_subscribe",
		"/system/subscribe", time.Minute)
	return runner.Start()
}
