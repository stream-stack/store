package subscriber

import "github.com/stream-stack/store/pkg/grpc"

func init() {
	//grpc.RegisterSubscriberFactory(func(kvsvc *grpc.KeyValueService, evsvc *grpc.EventService) grpc.Subscriber {
	//	return newIndexSubscriber(evsvc, kvsvc)
	//})
	grpc.RegisterSubscriberFactory(func(kvsvc *grpc.KeyValueService, evsvc *grpc.EventService) grpc.Subscriber {
		return newSlotDigestSubscriber(evsvc, kvsvc)
	})
}
