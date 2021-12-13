package grpc

import (
	"context"
	"fmt"
	"github.com/Jille/raft-grpc-leader-rpc/leaderhealth"
	"github.com/Jille/raftadmin"
	"github.com/stream-stack/store/pkg/config"
	"github.com/stream-stack/store/pkg/raft"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

func StartGrpc(ctx context.Context) error {
	_, port, err := net.SplitHostPort(config.Address)
	if err != nil {
		return fmt.Errorf("failed to parse local address (%q): %v", config.Address, err)
	}
	sock, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	raft.RaftManager.Register(s)
	leaderhealth.Setup(raft.Raft, s, []string{"example"})
	raftadmin.Register(s, raft.Raft)
	reflection.Register(s)
	go func() {
		select {
		case <-ctx.Done():
			s.GracefulStop()
		}
	}()
	if err := s.Serve(sock); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}
