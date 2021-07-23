package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stream-stack/store/pkg/config"
	"github.com/stream-stack/store/pkg/server"
	"github.com/stream-stack/store/pkg/storage"
	"github.com/stream-stack/store/pkg/storage/etcd"
	"os"
	"os/signal"
)

func NewCommand() (*cobra.Command, context.Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithCancel(context.TODO())
	c := &config.Config{}
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

			if err := viper.Unmarshal(c); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%+v \n", c)
			//1.storage
			if err := storage.Start(ctx, c); err != nil {
				return err
			}
			//5.grpc server
			if err := server.Start(ctx, c); err != nil {
				return err
			}

			<-ctx.Done()
			return nil
		},
	}
	viper.AutomaticEnv()
	viper.AddConfigPath(`.`)
	command.PersistentFlags().String("GrpcPort", "5001", "GrpcPort")
	command.PersistentFlags().String("StoreType", etcd.StoreType, "store type")
	command.PersistentFlags().StringSlice("StoreAddress", []string{"127.0.0.1:2379"}, "store address")
	command.PersistentFlags().String("PartitionType", storage.HASHPartitionType, "partition type")
	viper.SetDefault("GrpcPort", "5001")
	viper.SetDefault("StoreType", etcd.StoreType)
	viper.SetDefault("StoreAddress", "127.0.0.1:2379")
	viper.SetDefault("PartitionType", storage.HASHPartitionType)

	return command, ctx, cancelFunc
}

func main() {
	command, _, _ := NewCommand()
	if err := command.Execute(); err != nil {
		panic(err)
	}
}
