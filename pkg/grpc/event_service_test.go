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

func TestApply222(t *testing.T) {
	dial, err := grpc.Dial("localhost:2003", grpc.WithInsecure())
	if err != nil {
		panic(err.Error())
	}
	defer func() {
		_ = dial.Close()
	}()
	client := protocol.NewEventServiceClient(dial)

	todo := context.TODO()
	apply, err := client.Apply(todo, &protocol.ApplyRequest{
		StreamName: "a",
		StreamId:   "b",
		EventId:    "c",
		Data:       []byte(`test`),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(apply)
	get, err := client.Get(todo, &protocol.GetRequest{
		StreamName: "a",
		StreamId:   "b",
		EventId:    "c",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(get)
}

func TestApply(t *testing.T) {
	serviceConfig := `{"healthCheckConfig": {"serviceName": "example"}, "loadBalancingConfig": [ { "round_robin": {} } ]}`
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
		EventId:    "c",
		Data:       []byte(`test`),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(apply)
	get, err := client.Get(todo, &protocol.GetRequest{
		StreamName: "a",
		StreamId:   "b",
		EventId:    "c",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(get)
}
