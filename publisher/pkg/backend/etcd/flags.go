package etcd

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/store/store/common/config"
	"github.com/stream-stack/store/store/publisher/pkg/publisher"
	"time"
)

const BackendType = "ETCD"

func InitFlags() {
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
