package etcd

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/store/store/common/config"
	"github.com/stream-stack/store/store/pkg/storage"
	"time"
)

func InitFlags() {
	storage.Register(BackendType, NewStorageFunc)

	config.RegisterFlags(func(command *cobra.Command) {
		command.PersistentFlags().StringVar(&storage.BackendTypeValue, "BackendType", BackendType, "store type")

		command.PersistentFlags().StringVar(&Username, "EtcdUsername", "", "etcd Username")
		command.PersistentFlags().StringVar(&Password, "EtcdPassword", "", "etcd Password")
		command.PersistentFlags().DurationVar(&Timeout, "EtcdTimeout", time.Second*5, "etcd connect Timeout")
	})
}

var Username string
var Password string
var Timeout time.Duration
