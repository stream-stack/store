package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stream-stack/monitor/pkg/config"
	"github.com/stream-stack/monitor/pkg/watcher"
	"github.com/stream-stack/monitor/pkg/watcher/etcd"
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
			//watcher
			if err := watcher.Start(ctx, c); err != nil {
				return err
			}

			<-ctx.Done()
			return nil
		},
	}
	viper.AutomaticEnv()
	viper.AddConfigPath(`.`)
	command.PersistentFlags().String("StoreType", "ETCD", "store type")
	command.PersistentFlags().StringSlice("StoreAddress", []string{"127.0.0.1:2379"}, "store address")
	//command.PersistentFlags().String("PartitionType", "HASH", "partition type")
	viper.SetDefault("StoreType", etcd.StoreType)
	viper.SetDefault("StoreAddress", "127.0.0.1:2379")
	//viper.SetDefault("PartitionType", "HASH")

	return command, ctx, cancelFunc
}

func main() {
	command, _, _ := NewCommand()
	if err := command.Execute(); err != nil {
		panic(err)
	}
}
