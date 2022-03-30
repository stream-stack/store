package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stream-stack/store/pkg/config"
	"github.com/stream-stack/store/pkg/grpc"
	"github.com/stream-stack/store/pkg/index"
	"github.com/stream-stack/store/pkg/raft"
	"github.com/stream-stack/store/pkg/snapshot"
	"github.com/stream-stack/store/pkg/wal"
	"os"
	"os/signal"
)

func NewCommand() (*cobra.Command, context.Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithCancel(context.TODO())
	command := &cobra.Command{
		Use:   ``,
		Short: ``,
		Long:  ``,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			go func() {
				c := make(chan os.Signal, 1)
				signal.Notify(c, os.Kill)
				<-c
				cancelFunc()
			}()
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logrus.SetLevel(logrus.TraceLevel)
			if err := wal.StartWalEngine(ctx); err != nil {
				return err
			}
			if err := index.StartFSM(ctx); err != nil {
				return err
			}
			if err := raft.StartRaft(ctx); err != nil {
				return err
			}
			if err := grpc.StartGrpc(ctx); err != nil {
				return err
			}

			<-ctx.Done()
			return nil
		},
	}
	config.InitFlags()
	grpc.InitFlags()
	index.InitFlags()
	raft.InitFlags()
	snapshot.InitFlags()
	wal.InitFlags()

	viper.AutomaticEnv()
	viper.AddConfigPath(`.`)
	config.BuildFlags(command)

	return command, ctx, cancelFunc
}

func main() {
	command, _, _ := NewCommand()
	if err := command.Execute(); err != nil {
		panic(err)
	}
}
