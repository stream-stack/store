package grpc

import (
	"context"
	"fmt"
	_ "github.com/Jille/grpc-multi-resolver"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/stream-stack/store/pkg/protocol"
	"google.golang.org/grpc"
	"log"
	"testing"
	"time"
)

func TestApply(t *testing.T) {
	serviceConfig := `{"healthCheckConfig": {"serviceName": ""}, "loadBalancingConfig": [ { "round_robin": {} } ]}`
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
		grpc_retry.WithMax(5),
	}
	conn, err := grpc.Dial("multi:///localhost:2001,localhost:2002,localhost:2003",
		grpc.WithDefaultServiceConfig(serviceConfig), grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)))
	if err != nil {
		log.Fatalf("dialing failed: %v", err)
	}
	defer conn.Close()
	client := protocol.NewEventServiceClient(conn)

	todo := context.TODO()
	apply, err := client.Apply(todo, &protocol.ApplyRequest{
		StreamName: "a",
		StreamId:   "b",
		EventId:    "2",
		Data:       []byte(`test`),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(apply)
	get, err := client.Read(todo, &protocol.ReadRequest{
		StreamName: "a",
		StreamId:   "b",
		EventId:    "c",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(get)
}

func TestSubscribe(t *testing.T) {
	serviceConfig := `{"healthCheckConfig": {"serviceName": ""}, "loadBalancingConfig": [ { "round_robin": {} } ]}`
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
		grpc_retry.WithMax(5),
	}
	conn, err := grpc.Dial("multi:///localhost:2001,localhost:2002,localhost:2003",
		grpc.WithDefaultServiceConfig(serviceConfig), grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)))
	if err != nil {
		log.Fatalf("dialing failed: %v", err)
	}
	defer conn.Close()
	client := protocol.NewEventServiceClient(conn)

	todo, cancel := context.WithCancel(context.TODO())

	go func() {
		subscribe, err := client.Subscribe(todo, &protocol.SubscribeRequest{
			SubscribeId: "1",
			Regexp:      "",
			Offset:      0,
		})
		if err != nil {
			panic(err)
		}
		for {
			select {
			case <-todo.Done():
				return
			default:
				recv, err := subscribe.Recv()
				if err != nil {
					panic(err)
				}
				fmt.Println(recv)
			}
		}
	}()
	for i := 1; i < 10; i++ {
		apply, err := client.Apply(todo, &protocol.ApplyRequest{
			StreamName: "a",
			StreamId:   "b",
			EventId:    fmt.Sprintf("%d", i),
			Data:       []byte(fmt.Sprintf("%d-test", i)),
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(apply)
		time.Sleep(time.Second)
	}
	cancel()
}
