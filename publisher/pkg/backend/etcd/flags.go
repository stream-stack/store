package etcd

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/publisher/pkg/config"
	"github.com/stream-stack/publisher/pkg/publisher"
	"github.com/stream-stack/publisher/pkg/watcher"
	"time"
)

func InitFlags() {
	watcher.Register(BackendType, NewStorageFunc)
	publisher.Register(BackendType, NewSubscribeManagerFunc)

	config.RegisterFlags(func(command *cobra.Command) {
		command.PersistentFlags().StringVar(&publisher.BackendTypeValue, "BackendType", BackendType, "store type")

		command.PersistentFlags().StringVar(&Username, "EtcdUsername", "", "etcd Username")
		command.PersistentFlags().StringVar(&Password, "EtcdPassword", "", "etcd Password")
		command.PersistentFlags().DurationVar(&Timeout, "EtcdTimeout", time.Second*5, "etcd connect Timeout")
	})
}

var Username string
var Password string
var Timeout time.Duration
