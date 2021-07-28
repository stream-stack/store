package publisher

import (
	"context"
)

type SubscribeManager interface {
	LoadAllSubscribe(ctx context.Context) (SubscriberList, error)
}

type SubscribeManagerFactory func(ctx context.Context, storeAddress string) (SubscribeManager, error)

const LastEvent = "LAST"
const SubscribeStreamName = "subscribes"
const SubscribeStreamId = "_publisher"

type StartPoint interface {
	Compare(s StartPoint) int
}

type CurrentStartPoint struct {
}

func (c *CurrentStartPoint) Compare(_ StartPoint) int {
	return 0
}

type SubscriberList []*Subscriber

func (s SubscriberList) Len() int {
	return len(s)
}

func (s SubscriberList) Less(i, j int) bool {
	if s[i].StartPoint.Compare(s[j].StartPoint) > 0 {
		return true
	} else {
		return false
	}
}

func (s SubscriberList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
