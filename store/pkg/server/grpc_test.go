package server

import (
	"context"
	"fmt"
	"github.com/stream-stack/store/pkg/config"
	"github.com/stream-stack/store/pkg/proto"
	"github.com/stream-stack/store/pkg/storage"
	"github.com/stream-stack/store/pkg/storage/etcd"
	"google.golang.org/grpc"
	"log"
	"strconv"
	"testing"
)

func TestSave(t *testing.T) {
	port := "5051"
	c := &config.Config{
		GrpcPort:      port,
		StoreType:     etcd.StoreType,
		StoreAddress:  []string{"111.231.212.35:2379"},
		PartitionType: storage.HASHPartitionType,
	}
	todo := context.TODO()
	todo, cancelFunc := context.WithCancel(todo)
	//cluster
	if err := storage.Start(todo, c); err != nil {
		panic(err)
	}
	//grpc server
	if err := Start(todo, c); err != nil {
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
	for i := 300; i < 400; i++ {
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
	c := &config.Config{
		GrpcPort:      port,
		StoreType:     etcd.StoreType,
		StoreAddress:  []string{"111.231.212.35:2379"},
		PartitionType: storage.HASHPartitionType,
	}
	todo := context.TODO()
	todo, cancelFunc := context.WithCancel(todo)
	//cluster
	if err := storage.Start(todo, c); err != nil {
		panic(err)
	}
	//grpc server
	if err := Start(todo, c); err != nil {
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
		EventId:    storage.FirstEvent,
	})
	if err != nil {
		fmt.Println(err)
	}
	log.Printf("收到服务器回复:{%+v} \n", save)
	save, err = cli.Get(todo, &proto.GetRequest{
		StreamName: "test",
		StreamId:   "1",
		EventId:    storage.LastEvent,
	})
	if err != nil {
		fmt.Println(err)
	}
	log.Printf("收到服务器回复:{%+v} \n", save)
	cancelFunc()
	//abs, _ := filepath.Abs(c.DataDir)
	//os.RemoveAll(abs)
}
