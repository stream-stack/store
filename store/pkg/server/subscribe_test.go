package server

import (
	"context"
	"fmt"
	"github.com/stream-stack/store/store/common/proto"
	"github.com/stream-stack/store/store/pkg/storage"
	"github.com/stream-stack/store/store/pkg/storage/etcd"
	"google.golang.org/grpc"
	"log"
	"strconv"
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

func TestSubscribeReceive(t *testing.T) {
	port := "5001"
	GrpcPort = port

	storage.BackendAddressValue = []string{"127.0.0.1:2379"}
	storage.BackendTypeValue = etcd.BackendType

	etcd.InitFlags()
	etcd.Timeout = time.Second * 5

	todo := context.TODO()
	todo, cancelFunc := context.WithCancel(todo)

	dial, err := grpc.Dial("localhost:"+port, grpc.WithInsecure())
	if err != nil {
		panic(err.Error())
	}
	defer func() {
		_ = dial.Close()
	}()
	sbSvc := proto.NewSubscribeServiceClient(dial)
	create, err := sbSvc.Create(todo, &proto.CreateRequest{
		Name:                     "subscribe1",
		StartPoint:               proto.CreateRequest_Current,
		WatchKey:                 "/test",
		PushUrl:                  "http://localhost:8080",
		PushMaxRequestDuration:   "10s",
		PushMaxRequestRetryCount: 10,
		SaveInterval:             "100s",
	})
	if err != nil {
		panic(err)
	}
	log.Printf("创建订阅完成,返回值:%v", create)
	eSvc := proto.NewEventServiceClient(dial)
	for i := 1; i < 5; i++ {
		save, err := eSvc.Save(todo, &proto.SaveRequest{
			StreamName: "test",
			StreamId:   "1",
			EventId:    strconv.Itoa(i),
			Data:       []byte(fmt.Sprintf("test-%d", i)),
		})
		if err != nil {
			fmt.Println(err)
		}
		log.Printf("收到服务器回复:{%+v} \n", save)
	}

	//response, err := sbSvc.Delete(todo, &proto.DeleteRequest{
	//	Name: "subscribe1",
	//})
	//if err != nil {
	//	panic(err)
	//}
	//log.Printf("删除订阅完成,返回值:%v", response)
	for i := 1; i < 5; i++ {
		save, err := eSvc.Save(todo, &proto.SaveRequest{
			StreamName: "test",
			StreamId:   "2",
			EventId:    strconv.Itoa(i),
			Data:       []byte(fmt.Sprintf("test-%d", i)),
		})
		if err != nil {
			fmt.Println(err)
		}
		log.Printf("收到服务器回复:{%+v} \n", save)
	}
	cancelFunc()
}
