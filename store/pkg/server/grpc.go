package server

import (
	"context"
	"fmt"
	"github.com/stream-stack/store/store/common/proto"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
)

func Start(ctx context.Context) error {
	grpcServer := grpc.NewServer()
	es := &EventService{}
	proto.RegisterEventServiceServer(grpcServer, es)
	proto.RegisterSubscribeServiceServer(grpcServer, &SubscribeService{eventService: es})

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", GrpcPort))
		if err != nil {
			log.Panicf("[grpc]listen error:%v", err)
		}
		if err := grpcServer.Serve(lis); err != nil {
			log.Panicf("[grpc]failed to serve: %v", err)
		}
	}()
	go func() {
		err := http.ListenAndServe("0.0.0.0:6060", nil)
		if err != nil {
			log.Printf("[pprof]start err:%v", err.Error())
		}
	}()
	go func() {
		select {
		case <-ctx.Done():
			grpcServer.GracefulStop()
		}
	}()
	return nil
}
