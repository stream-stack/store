package publisher

import (
	"context"
	"fmt"
)

type SubscribeManager interface {
	LoadAllSubscribe(ctx context.Context) (SubscriberList, error)
	LoadStreamSnapshot(ctx context.Context, streamName string, StreamId string) ([]byte, error)
}

type SubscribeManagerFactory func(ctx context.Context, storeAddress string) (SubscribeManager, error)

const LastEvent = "LAST"
const SubscribeStreamName = "subscribes"
const SubscribeStreamId = "_publisher"

type StartPoint interface {
	//0相等，1大于，-1 小于
	Compare(s StartPoint) int
}

type MapStartPoint map[string]uint64

func (m MapStartPoint) Compare(s StartPoint) int {

}

type UIntStartPoint struct {
	Point uint64 `json:"point"`
}

func (u UIntStartPoint) Compare(s StartPoint) int {
	point, ok := s.(UIntStartPoint)
	if !ok {
		panic(fmt.Errorf("startpoint %+v not type UIntStartPoint", s))
	}
	if u.Point > point.Point {
		return 1
	}
	if u.Point < point.Point {
		return -1
	}
	return 0
}

type BeginStartPoint struct {
}

func (b *BeginStartPoint) Compare(s StartPoint) int {
	return -1
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
