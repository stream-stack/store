package etcd

import (
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/stream-stack/store/store/publisher/pkg/publisher"
)

type etcdSnapshot struct {
	//快照创建位置
	StartPoint uint64 `json:"start_point"`
	//扩展数据
	ExtData []byte `json:"ext_data"`
}

func (e *etcdSnapshot) SetStartPoint(i publisher.StartPoint) {
	u := i.(uint64)
	e.StartPoint = u
}

func (e *etcdSnapshot) GetStartPoint() publisher.StartPoint {
	return e.StartPoint
}

func (e *etcdSnapshot) GetExtData() []byte {
	return e.ExtData
}

type etcdEvent struct {
	event.Event
	StartPoint uint64
}

func (e *etcdEvent) GetStartPoint() publisher.StartPoint {
	return e.StartPoint
}
