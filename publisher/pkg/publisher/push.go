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
		log.Printf("[subscribeOperation]收到创建subscribe请求,开始创建")
		ed[operation.Name] = operation.BaseSubscribe
		runner := NewSubscribeRunnerWithSubscribeOperation(runner, operation.BaseSubscribe)
		return runner.Start()
	case proto.Delete:
		log.Printf("[subscribeOperation]收到删除subscribe请求,开始删除")
		_, ok := ed[operation.Name]
		if !ok {
			return nil
		}
		delete(ed, operation.Name)
		runner, ok := runners[operation.Name]
		if !ok {
			return nil
		}
		runner.Stop()
	}
	runner.ExtData = ed
	return nil
}
