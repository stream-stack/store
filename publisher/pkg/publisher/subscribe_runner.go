package publisher

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

var runners = make(map[string]*SubscribeRunner)

type SubscribeRunner struct {
	ctx                                  context.Context
	ss                                   SubscribeManager
	dataUnmarshaler                      DataUnmarshaler
	eventUnmarshaler                     EventUnmarshaler
	name, streamName, streamId, watchKey string
	action                               action
	cancelFunc                           context.CancelFunc
	subCtx                               context.Context
	pushSetting                          SubscribePushSetting
	saveInterval                         time.Duration
}

func NewSubscribeRunner(ctx context.Context, ss SubscribeManager, dataUnmarshaler DataUnmarshaler, eventUnmarshaler EventUnmarshaler, name string,
	streamName string, streamId string, key string, action action, saveInterval time.Duration) *SubscribeRunner {
	return &SubscribeRunner{ctx: ctx, ss: ss, dataUnmarshaler: dataUnmarshaler, eventUnmarshaler: eventUnmarshaler,
		name: name, streamName: streamName, streamId: streamId, watchKey: key, action: action, saveInterval: saveInterval}
}
func NewSubscribeRunnerWithSubscribeOperation(s *BaseSubscribe) *SubscribeRunner {
	return &SubscribeRunner{ctx: s.ctx, ss: s.subscribeManager,
		dataUnmarshaler: du, eventUnmarshaler: eu,
		//"snapshot", "xxx"
		name: s.Name, streamName: "snapshot", streamId: s.Name,
		watchKey: s.Key, action: pushCloudEvent, pushSetting: s.SubscribePushSetting, saveInterval: s.Interval}
}

func (r *SubscribeRunner) start() error {
	subCtx, cancelFunc := context.WithCancel(r.ctx)
	r.cancelFunc = cancelFunc
	r.subCtx = subCtx
	sn, err := r.ss.LoadSnapshot(subCtx, r.streamName, r.streamId)
	if err != nil {
		return err
	}
	data, err := r.dataUnmarshaler(sn.GetExtData())

	watch, err := r.ss.Watch(subCtx, sn.GetStartPoint(), r.watchKey, r.eventUnmarshaler)
	if err != nil {
		return err
	}
	go func() {
		runners[r.name] = r
		defer func() {
			_, ok := runners[r.name]
			if !ok {
				return
			}
			delete(runners, r.name)
		}()
		for {
			select {
			case <-subCtx.Done():
				return
			case event := <-watch:
				sn.SetStartPoint(event.GetStartPoint())
				if err := r.action(r.subCtx, event, data, r.pushSetting); err != nil {
					log.Printf("[subscribe-runner]runner %v do action error:%v", r.watchKey, err)
				}
			}
		}
	}()
	go func() {
		//TODO:满足多个条件才保存,减少快照数量
		ticker := time.NewTicker(r.saveInterval)
		for {
			select {
			case <-subCtx.Done():
				return
			case <-ticker.C:
				marshal, err := json.Marshal(sn)
				if err != nil {
					log.Printf("[subscribe-runner]runner %v marshal snapshot error:%v", r.watchKey, err)
					continue
				}
				err = ss.SaveSnapshot(subCtx, r.streamName, r.streamId, marshal)
				if err != nil {
					log.Printf("[subscribe-runner]runner %v save snapshot data error:%v", r.watchKey, err)
				}
			}
		}
	}()
	return err
}

func (r *SubscribeRunner) stop() {
	r.cancelFunc()
}
