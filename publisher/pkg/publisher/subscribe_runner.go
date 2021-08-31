package publisher

import (
	"context"
	"log"
	"time"
)

var runners = make(map[string]*SubscribeRunner)

type SubscribeRunner struct {
	ctx                                  context.Context
	cancelFunc                           context.CancelFunc
	ss                                   SubscribeManager
	name, StreamName, StreamId, WatchKey string
	PushSetting                          SubscribePushSetting
	saveInterval                         time.Duration

	Action     Action
	ExtData    interface{}
	StartPoint StartPoint
}

func newSystemSubscribeRunner(ctx context.Context, ss SubscribeManager, name string,
	streamName string, streamId string, key string, saveInterval time.Duration) *SubscribeRunner {
	subCtx, cancelFunc := context.WithCancel(ctx)
	return &SubscribeRunner{
		ctx:          subCtx,
		cancelFunc:   cancelFunc,
		ss:           ss,
		name:         name,
		StreamName:   streamName,
		StreamId:     streamId,
		WatchKey:     key,
		saveInterval: saveInterval,
		Action:       subscribeOperation,
	}
}
func NewSubscribeRunnerWithSubscribeOperation(parent *SubscribeRunner, s *BaseSubscribe) *SubscribeRunner {
	subCtx, cancelFunc := context.WithCancel(parent.ctx)
	return &SubscribeRunner{
		ctx:        subCtx,
		cancelFunc: cancelFunc,
		ss:         parent.ss,
		//"snapshot", "xxx"
		name:         s.Name,
		StreamName:   "snapshot",
		StreamId:     s.Name,
		WatchKey:     s.Key,
		PushSetting:  s.SubscribePushSetting,
		saveInterval: s.Interval,
		Action:       pushCloudEvent,
	}
}

func (r *SubscribeRunner) Start() error {
	err := r.ss.LoadSnapshot(r.ctx, r)
	if err != nil {
		return err
	}

	go func() {
		err := r.ss.Watch(r.ctx, r)
		if err != nil {
			log.Printf("[subscribe-runner]runner %v watch error:%v", r.name, err)
		}
		runners[r.name] = r
		defer func() {
			_, ok := runners[r.name]
			if !ok {
				return
			}
			delete(runners, r.name)
		}()
	}()
	go func() {
		//TODO:满足多个条件才保存,减少快照数量
		ticker := time.NewTicker(r.saveInterval)
		for {
			select {
			case <-r.ctx.Done():
				return
			case <-ticker.C:
				err = ss.SaveSnapshot(r.ctx, r)
				if err != nil {
					log.Printf("[subscribe-runner]runner %v save snapshot data error:%v", r.name, err)
				}
			}
		}
	}()
	return err
}

func (r *SubscribeRunner) Stop() {
	r.cancelFunc()
}
