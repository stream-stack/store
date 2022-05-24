package grpc

import (
	"context"
	"fmt"
	_ "github.com/Jille/grpc-multi-resolver"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	protocol "github.com/stream-stack/common/protocol/store"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/health"
	"google.golang.org/grpc/status"
	"log"
	"testing"
	"time"
)

func eventExecutor(f func(ctx context.Context, client protocol.EventServiceClient)) {
	todo := context.TODO()
	serviceConfig := `{"healthCheckConfig": {"serviceName": "store"}, "loadBalancingConfig": [ { "round_robin": {} } ]}`
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
	f(todo, client)
}

func TestExistEventId(t *testing.T) {
	//send exist event
	eventExecutor(func(ctx context.Context, client protocol.EventServiceClient) {
		apply, err := client.Apply(ctx, &protocol.ApplyRequest{
			StreamName: "a",
			StreamId:   "b",
			EventId:    2,
			Data:       []byte(`test`),
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(apply)
	})
	eventExecutor(func(ctx context.Context, client protocol.EventServiceClient) {
		apply, err := client.Apply(ctx, &protocol.ApplyRequest{
			StreamName: "a",
			StreamId:   "b",
			EventId:    2,
			Data:       []byte(`test`),
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(apply)
	})
}

func TestApply(t *testing.T) {
	eventExecutor(func(ctx context.Context, client protocol.EventServiceClient) {
		read, err := client.Read(ctx, &protocol.ReadRequest{
			StreamName: "a",
			StreamId:   "b",
			EventId:    2,
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(read)
	})
}

func TestSubscribe(t *testing.T) {
	serviceConfig := `{"healthCheckConfig": {"serviceName": "store"}, "loadBalancingConfig": [ { "round_robin": {} } ]}`
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
			Regexp:      "streamName=~ '[a-z]+' ",
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
	for i := 1; i < 1000000; i++ {
		apply, err := client.Apply(todo, &protocol.ApplyRequest{
			StreamName: "a",
			StreamId:   "b",
			EventId:    uint64(i),
			Data:       []byte(fmt.Sprintf("%d-test", i)),
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(apply)
		time.Sleep(time.Second * 5)
	}
	time.Sleep(time.Hour)
	cancel()
}

func TestPut(t *testing.T) {
	serviceConfig := `{"healthCheckConfig": {"serviceName": "store"}, "loadBalancingConfig": [ { "round_robin": {} } ]}`
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
	kvClient := protocol.NewKVServiceClient(conn)
	ctx := context.TODO()
	put, err := kvClient.Put(ctx, &protocol.PutRequest{
		Key: []byte(`test`),
		Val: protocol.Uint64ToBytes(2),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(put)
}

func TestGet(t *testing.T) {
	serviceConfig := `{"healthCheckConfig": {"serviceName": "store"}, "loadBalancingConfig": [ { "round_robin": {} } ]}`
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
	kvClient := protocol.NewKVServiceClient(conn)
	ctx := context.TODO()
	put, err := kvClient.Get(ctx, &protocol.GetRequest{
		Key: []byte(`test`),
	})
	if err != nil {
		convert := status.Convert(err)
		fmt.Println(convert)
		panic(err)
	}
	fmt.Println(put)
}
