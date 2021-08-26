package publisher

import (
	"context"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"log"
)

func pushCloudEvent(ctx context.Context, event SubscribeEvent, data interface{}, push SubscribePushSetting) error {
	ctx = cloudevents.ContextWithTarget(ctx, push.Url)
	ctx = cloudevents.ContextWithRetriesExponentialBackoff(ctx, push.MaxRequestDuration, push.MaxRequestRetryCount)

	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
		return err
	}

	c, err := cloudevents.NewClient(p, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	res := c.Send(ctx, event.GetCloudEvent())
	if cloudevents.IsUndelivered(res) {
		log.Printf("Failed to send: %v", res)
		return fmt.Errorf("[cloudevent]Failed to send: %v", res)
	} else {
		var httpResult *cehttp.Result
		cloudevents.ResultAs(res, &httpResult)
		log.Printf("Sent data %+v with status code %d", data, httpResult.StatusCode)
	}

	return nil
}
