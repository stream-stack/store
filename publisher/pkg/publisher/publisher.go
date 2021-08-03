package publisher

import (
	"context"
	"encoding/json"
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
	//加载overview stream
	data, err := ss.LoadStreamSnapshot(ctx, "_System", "stream_overview")
	if err != nil {
		return err
	}
	systemSubscribe, err := newSystemStreamOverview(data)
	if err != nil {
		return err
	}
	subscribers = append(subscribers, systemSubscribe)
	if len(subscribers) == 0 {
		WatchStartPoint = &CurrentStartPoint{}
	} else {
		sort.Sort(subscribers)
		WatchStartPoint = subscribers[0].StartPoint
	}

	startPublisher(ctx)
	return nil
}

func newSystemStreamOverview(data []byte) (*Subscriber, error) {
	if data == nil {
		return newStreamOverview(&BeginStartPoint{}), nil
	}
	overview := &StreamOverview{}
	if err := json.Unmarshal(data, overview); err != nil {
		return newStreamOverview(&BeginStartPoint{}), err
	}
	return newStreamOverview(overview.minPoint()), nil
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
						val.DoAction(e)
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
