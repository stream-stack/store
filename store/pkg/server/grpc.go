package server

import (
	"context"
	"fmt"
	"github.com/stream-stack/store/pkg/proto"
	"github.com/stream-stack/store/pkg/storage"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
)

type StorageService struct {
}

func (s *StorageService) Get(ctx context.Context, request *proto.GetRequest) (*proto.GetResponse, error) {
	data, err := storage.SaveStorage.Get(ctx, request.StreamName, request.StreamId, request.EventId)
	if err != nil {
		return &proto.GetResponse{
			Message: err.Error(),
		}, err
	}
	return &proto.GetResponse{
		EventId: request.EventId,
		Data:    data,
	}, nil
}

func (s *StorageService) Save(ctx context.Context, request *proto.SaveRequest) (*proto.SaveResponse, error) {
	log.Printf("[grpc]handler save request{%+v}", request)

	if err := storage.SaveStorage.Save(ctx, request.StreamName, request.StreamId, request.EventId, request.Data); err != nil {
		return &proto.SaveResponse{
			Ack:     false,
			Message: err.Error(),
		}, err
	}
	return &proto.SaveResponse{
		Ack: true,
	}, nil
}

func Start(ctx context.Context) error {
	grpcServer := grpc.NewServer()
	proto.RegisterStorageServer(grpcServer, &StorageService{})

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
