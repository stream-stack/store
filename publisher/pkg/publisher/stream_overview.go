package publisher

import "github.com/stream-stack/publisher/pkg/proto"

type SystemStreamOverview struct {
	Points map[string]UIntStartPoint `json:"points"`
}

func (o *SystemStreamOverview) minPoint() StartPoint {
	var minPoint StartPoint
	for _, point := range o.Points {
		if minPoint == nil {
			minPoint = point
			continue
		}
		if point.Compare(minPoint) > 0 {
			minPoint = point
		}
	}
	return minPoint
}

func newStreamOverview(point StartPoint) *Subscriber {
	return &Subscriber{
		Name:       "_System_stream_overview",
		Pattern:    "*",
		StartPoint: point,
		action: func(subscriber *Subscriber, request proto.SaveRequest) error {
			//todo:执行存储操作
			return nil
		},
	}
}
