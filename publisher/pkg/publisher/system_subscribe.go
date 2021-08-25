package publisher

import (
	"context"
	"encoding/json"
)

func systemDataUnmarshaler(data []byte) (interface{}, error) {
	subscribes := make(map[string]*BaseSubscribe)
	err := json.Unmarshal(data, &subscribes)
	if err != nil {
		return nil, err
	}
	for _, sb := range subscribes {
		runner := NewSubscribeRunnerWithSubscribeOperation(sb)
		if err := runner.start(); err != nil {
			return nil, err
		}
	}
	return subscribes, nil
}

func systemAction(ctx context.Context, event SubscribeEvent, data interface{}, push SubscribePushSetting) error {
	operation := event.(SubscribeOperation)
	operation.ctx = ctx
	operation.subscribeManager = ss
	return operation.update(data.(map[string]*BaseSubscribe))
}

type SubscribeOperation struct {
	//创建,删除订阅
	Operation string `json:"operation"`
	//基础信息
	*BaseSubscribe `json:"-"`
}

func (o SubscribeOperation) update(subscribes map[string]*BaseSubscribe) error {
	switch o.Operation {
	case "create":
		subscribes[o.Name] = o.BaseSubscribe
		runner := NewSubscribeRunnerWithSubscribeOperation(o.BaseSubscribe)
		return runner.start()
	case "delete":
		_, ok := subscribes[o.Name]
		if !ok {
			return nil
		}
		delete(subscribes, o.Name)
		runner, ok := runners[o.Name]
		if !ok {
			return nil
		}
		runner.stop()
	}
	return nil
}
