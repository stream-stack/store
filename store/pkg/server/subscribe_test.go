package server

import (
	"context"
	"fmt"
	"github.com/stream-stack/store/store/common/proto"
	"github.com/stream-stack/store/store/pkg/storage"
	"github.com/stream-stack/store/store/pkg/storage/etcd"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestSubscribeCreate(t *testing.T) {
	port := "5001"
	GrpcPort = port

	storage.BackendAddressValue = []string{"127.0.0.1:2379"}
	storage.BackendTypeValue = etcd.BackendType

	etcd.InitFlags()
	//etcd.Username = ""
	//etcd.Password = ""
	etcd.Timeout = time.Second * 5

	todo := context.TODO()
	todo, cancelFunc := context.WithCancel(todo)
	//cluster
	//if err := storage.Start(todo); err != nil {
	//	panic(err)
	//}
	//grpc server
	//if err := Start(todo); err != nil {
	//	panic(err)
	//}

	dial, err := grpc.Dial("localhost:"+port, grpc.WithInsecure())
	if err != nil {
		panic(err.Error())
	}
	defer func() {
		_ = dial.Close()
	}()
	client := proto.NewSubscribeServiceClient(dial)
	create, err := client.Create(todo, &proto.CreateRequest{
		Name:                     "test1",
		StartPoint:               proto.CreateRequest_Current,
		WatchKey:                 "/test",
		PushUrl:                  "http://www.baidu.com",
		PushMaxRequestDuration:   "10s",
		PushMaxRequestRetryCount: 10,
		SaveInterval:             "100s",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(create)
	cancelFunc()
}

func TestSubscribeDelete(t *testing.T) {
	port := "5001"
	GrpcPort = port

	storage.BackendAddressValue = []string{"127.0.0.1:2379"}
	storage.BackendTypeValue = etcd.BackendType

	etcd.InitFlags()
	//etcd.Username = ""
	//etcd.Password = ""
	etcd.Timeout = time.Second * 5

	todo := context.TODO()
	todo, cancelFunc := context.WithCancel(todo)
	//cluster
	//if err := storage.Start(todo); err != nil {
	//	panic(err)
	//}
	//grpc server
	//if err := Start(todo); err != nil {
	//	panic(err)
	//}

	dial, err := grpc.Dial("localhost:"+port, grpc.WithInsecure())
	if err != nil {
		panic(err.Error())
	}
	defer func() {
		_ = dial.Close()
	}()
	client := proto.NewSubscribeServiceClient(dial)
	response, err := client.Delete(todo, &proto.DeleteRequest{
		Name: "test1",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(response)
	cancelFunc()
}
