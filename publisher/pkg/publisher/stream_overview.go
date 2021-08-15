package publisher

import "github.com/stream-stack/publisher/pkg/proto"

func newStreamOverview(point StartPoint) *Subscriber {
	return &Subscriber{
		Name:       "_system_stream_overview",
		Pattern:    "*",
		StartPoint: point,
		action: func(s *Subscriber, request proto.SaveRequest) error {
			//todo:执行存储操作
			//更新startpoint，存储startpoint
			s.StartPoint.Update(request)
			return nil
		},
	}
}
