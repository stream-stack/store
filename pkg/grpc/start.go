package grpc

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stream-stack/store/pkg/cloudevents.io/genproto/v1"
	"google.golang.org/grpc"
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
	v1.RegisterStoreServer(s, NewStoreService())
	//marshal, err := proto.Marshal(&v1.CloudEvent{})
	//protocol.RegisterKVServiceServer(s, NewKVService())
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
