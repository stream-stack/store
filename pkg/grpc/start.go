package grpc

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	v1 "github.com/stream-stack/common/cloudevents.io/genproto/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"net"
)

func StartGrpc(ctx context.Context) error {
	Port := viper.GetString("Port")
	logrus.Debugf("[grpc]grpc starting at port:%v", Port)
	sock, err := net.Listen("tcp", fmt.Sprintf(":%s", Port))
	if err != nil {
		return fmt.Errorf("[grpc]failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	keyValueService := newPrivateKeyValueServiceServer()
	eventService := newPublicEventServiceServer()
	eventService.startSubscribeServer(ctx)
	if err := eventService.startIndexServer(ctx, keyValueService); err != nil {
		return err
	}
	v1.RegisterPrivateKeyValueServiceServer(s, keyValueService)
	v1.RegisterPublicEventServiceServer(s, eventService)

	hsrv := health.NewServer()
	hsrv.SetServingStatus("store", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(s, hsrv)
	go func() {
		select {
		case <-ctx.Done():
			s.GracefulStop()
		}
	}()
	if err := s.Serve(sock); err != nil {
		return fmt.Errorf("[grpc]failed to serve: %v", err)
	}
	return nil
}
