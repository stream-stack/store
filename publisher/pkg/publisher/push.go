package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/stream-stack/store/store/common/proto"
	"log"
)

func pushCloudEvent(ctx context.Context, event cloudevents.Event, runner *SubscribeRunner) error {
	log.Printf("[event-push]开始推送消息至[%s]", runner.PushSetting.Url)
	ctx = cloudevents.ContextWithTarget(ctx, runner.PushSetting.Url)
	ctx = cloudevents.ContextWithRetriesExponentialBackoff(ctx, runner.PushSetting.MaxRequestDuration, runner.PushSetting.MaxRequestRetryCount)

	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
		return err
	}

	c, err := cloudevents.NewClient(p, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	res := c.Send(ctx, event)
	log.Printf("[event-push]目标[%s]推送结果:%+v", runner.PushSetting.Url, res)
	if cloudevents.IsUndelivered(res) {
		log.Printf("Failed to send: %v", res)
		return fmt.Errorf("[cloudevent]Failed to send: %v", res)
	} else {
		var httpResult *cehttp.Result
		cloudevents.ResultAs(res, &httpResult)
		log.Printf("Sent data %+v with status code %+v", event, httpResult)
	}

	return nil
}

func subscribeOperation(ctx context.Context, event cloudevents.Event, runner *SubscribeRunner) error {
	log.Printf("[subscribeOperation]收到创建/删除subscribe请求,event:%v", event)
	operation := &proto.SubscribeOperation{}
	err := json.Unmarshal(event.Data(), operation)
	if err != nil {
		return err
	}
	ed := runner.ExtData.(map[string]*proto.BaseSubscribe)
	switch operation.Operation {
	case proto.Create:
		log.Printf("[subscribeOperation]收到创建subscribe[name=%s]请求,开始创建", operation.Name)
		ed[operation.Name] = operation.BaseSubscribe
		runner := NewSubscribeRunnerWithSubscribeOperation(runner, operation.BaseSubscribe)
		return runner.Start()
	case proto.Delete:
		log.Printf("[subscribeOperation]收到删除subscribe[name=%s]请求,开始删除", operation.Name)
		runners.Range(func(key, value interface{}) bool {
			log.Printf("==============runners key:%v,value:%v", key, value)
			return true
		})
		runner, ok := runners.Load(operation.Name)
		if !ok {
			log.Printf("[subscribeOperation]在runners中未找到key:%s", operation.Name)
		} else {
			subscribeRunner := runner.(*SubscribeRunner)
			subscribeRunner.Stop()
		}
		_, ok = ed[operation.Name]
		if !ok {
			log.Printf("[subscribeOperation]在map中未找到key:%s", operation.Name)
		} else {
			delete(ed, operation.Name)
		}
		log.Printf("[subscribeOperation]停止完成")
		log.Printf("当前ed:%+v", ed)
		for s, subscribe := range ed {
			log.Printf("当前ed key:%+v,value:%+v", s, subscribe)
		}

	}
	runner.ExtData = ed
	return nil
}
