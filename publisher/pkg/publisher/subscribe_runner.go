package publisher

import (
	"context"
	"github.com/stream-stack/store/store/common/proto"
	"github.com/stream-stack/store/store/common/vars"
	"log"
	"sync"
	"time"
)

var runners = sync.Map{}

type SubscribeRunner struct {
	ctx                                  context.Context
	cancelFunc                           context.CancelFunc
	ss                                   SubscribeManager
	name, StreamName, StreamId, WatchKey string
	PushSetting                          proto.SubscribePushSetting
	saveInterval                         time.Duration

	Action Action
	//TODO:针对数据更新做并发控制
	ExtData    interface{}
	StartPoint proto.StartPoint
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
		WatchKey:     vars.StorePrefix + key,
		saveInterval: saveInterval,
		Action:       subscribeOperation,
	}
}
func NewSubscribeRunnerWithSubscribeOperation(parent *SubscribeRunner, s *proto.BaseSubscribe) *SubscribeRunner {
	subCtx, cancelFunc := context.WithCancel(parent.ctx)
	return &SubscribeRunner{
		ctx:        subCtx,
		cancelFunc: cancelFunc,
		ss:         parent.ss,
		//"snapshot", "xxx"
		name:         s.Name,
		StreamName:   "snapshot",
		StreamId:     s.Name,
		WatchKey:     vars.StorePrefix + s.Key,
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
		runners.Store(r.name, r)
		defer func() {
			runners.Delete(r.name)
			r.cancelFunc()
		}()
		err := r.ss.Watch(r.ctx, r)
		if err != nil {
			log.Printf("[subscribe-runner]runner %v watch error:%v", r.name, err)
		}
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
