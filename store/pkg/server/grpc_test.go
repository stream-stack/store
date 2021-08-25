package server

import (
	"context"
	"fmt"
	"github.com/stream-stack/store/store/common/proto"
	"github.com/stream-stack/store/store/common/vars"
	"github.com/stream-stack/store/store/pkg/storage"
	"github.com/stream-stack/store/store/pkg/storage/etcd"
	"google.golang.org/grpc"
	"log"
	"strconv"
	"testing"
	"time"
)

func TestSave(t *testing.T) {
	port := "5001"
	GrpcPort = port

	storage.BackendAddressValue = []string{"139.155.161.172:2379"}
	storage.BackendTypeValue = etcd.BackendType

	etcd.InitFlags()
	etcd.Username = "root"
	etcd.Password = "i6n629GqcE"
	etcd.Timeout = time.Second * 5

	todo := context.TODO()
	todo, cancelFunc := context.WithCancel(todo)
	//cluster
	if err := storage.Start(todo); err != nil {
		panic(err)
	}
	//grpc server
	if err := Start(todo); err != nil {
		panic(err)
	}

	dial, err := grpc.Dial("localhost:"+port, grpc.WithInsecure())
	if err != nil {
		panic(err.Error())
	}
	defer func() {
		_ = dial.Close()
	}()
	cli := proto.NewStorageClient(dial)
	for i := 1; i < 10; i++ {
		save, err := cli.Save(todo, &proto.SaveRequest{
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
	cancelFunc()
	//abs, _ := filepath.Abs(c.DataDir)
	//os.RemoveAll(abs)
}

func TestGet(t *testing.T) {
	port := "5051"
	GrpcPort = port

	storage.BackendAddressValue = []string{"127.0.0.1:2379"}
	storage.BackendTypeValue = etcd.BackendType

	etcd.InitFlags()
	etcd.Username = "root"
	etcd.Password = ""
	etcd.Timeout = time.Second * 5

	todo := context.TODO()
	todo, cancelFunc := context.WithCancel(todo)
	//cluster
	if err := storage.Start(todo); err != nil {
		panic(err)
	}
	//grpc server
	if err := Start(todo); err != nil {
		panic(err)
	}

	dial, err := grpc.Dial("localhost:"+port, grpc.WithInsecure())
	if err != nil {
		panic(err.Error())
	}
	defer func() {
		_ = dial.Close()
	}()
	cli := proto.NewStorageClient(dial)
	for i := 1; i < 5; i++ {
		save, err := cli.Save(todo, &proto.SaveRequest{
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
	save, err := cli.Get(todo, &proto.GetRequest{
		StreamName: "test",
		StreamId:   "1",
		EventId:    strconv.Itoa(6),
	})
	if err != nil {
		fmt.Println(err)
	}
	log.Printf("收到服务器回复:{%+v} \n", save)

	save, err = cli.Get(todo, &proto.GetRequest{
		StreamName: "test",
		StreamId:   "1",
		EventId:    vars.FirstEvent,
	})
	if err != nil {
		fmt.Println(err)
	}
	log.Printf("收到服务器回复:{%+v} \n", save)
	save, err = cli.Get(todo, &proto.GetRequest{
		StreamName: "test",
		StreamId:   "1",
		EventId:    vars.LastEvent,
	})
	if err != nil {
		fmt.Println(err)
	}
	log.Printf("收到服务器回复:{%+v} \n", save)
	cancelFunc()
	//abs, _ := filepath.Abs(c.DataDir)
	//os.RemoveAll(abs)
}
