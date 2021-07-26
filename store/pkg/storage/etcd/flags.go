package etcd

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/store/pkg/config"
	"github.com/stream-stack/store/pkg/storage"
	"time"
)

func InitFlags() {
	storage.Register(StoreType, NewStorageFunc)

	config.RegisterFlags(func(command *cobra.Command) {
		command.PersistentFlags().StringVar(&storage.StoreTypeValue, "StoreType", StoreType, "store type")
		command.PersistentFlags().StringSliceVar(&storage.StoreAddressValue, "StoreAddress", []string{"127.0.0.1:2379"}, "store address")

		command.PersistentFlags().StringVar(&Username, "EtcdUsername", "", "etcd Username")
		command.PersistentFlags().StringVar(&Password, "EtcdPassword", "", "etcd Password")
		command.PersistentFlags().DurationVar(&Timeout, "EtcdTimeout", time.Second*5, "etcd connect Timeout")
	})
}

var Username string
var Password string
var Timeout time.Duration
