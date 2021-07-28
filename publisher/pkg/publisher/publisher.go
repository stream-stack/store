package publisher

import (
	"context"
	"fmt"
	"github.com/stream-stack/publisher/pkg/proto"
	"log"
	"sort"
)

type BackendType string

var factoryMap = make(map[BackendType]SubscribeManagerFactory)

func Register(t BackendType, f SubscribeManagerFactory) {
	factoryMap[t] = f
}

var WatchStartPoint StartPoint
var subscribers SubscriberList

//TODO:订阅快照实现
func Start(ctx context.Context) error {
	factory, ok := factoryMap[BackendType(BackendTypeValue)]
	if !ok {
		return fmt.Errorf("[publisher]not support store type :%v", BackendTypeValue)
	}
	ss, err := factory(ctx, StoreAddress)
	if err != nil {
		return fmt.Errorf("[publisher]init storage error:%v", err)
	}
	subscribers, err = ss.LoadAllSubscribe(ctx)
	if err != nil {
		return err
	}
	for _, subscriber := range subscribers {
		err = subscriber.init()
		if err != nil {
			log.Printf("[publisher]subscribe init error:%v", err)
			continue
		}
	}
	if len(subscribers) == 0 {
		WatchStartPoint = &CurrentStartPoint{}
	} else {
		sort.Sort(subscribers)
		WatchStartPoint = subscribers[0].StartPoint
	}

	startPublisher(ctx)
	return nil
}

var In chan proto.SaveRequest

func startPublisher(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-In:
				for _, val := range subscribers {
					//go func(val *SubscribeEvent) {
					if val.Match(GetKey(e)) {
						val.Send(e)
					}
					//}(val)
				}
			}
		}
	}()
}

func GetKey(e proto.SaveRequest) string {
	return fmt.Sprintf("%s/%s/%s", e.StreamName, e.StreamId, e.EventId)
}
