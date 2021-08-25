package etcd

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/spf13/cobra"
	"github.com/stream-stack/store/store/common/config"
	"github.com/stream-stack/store/store/publisher/pkg/publisher"
	"time"
)

const BackendType = "ETCD"

func InitFlags() {
	publisher.Register(BackendType, NewSubscribeManagerFunc)
	publisher.RegisterExtDataUnmarshal(BackendType, NoneExtDataUnmarshaler)
	publisher.RegisterEventUnmarshal(BackendType, cloudEventUnmarshaler)

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

func NoneExtDataUnmarshaler(data []byte) (interface{}, error) {
	return nil, nil
}

func cloudEventUnmarshaler(key string, event interface{}) ([]publisher.SubscribeEvent, error) {
	w := event.(clientv3.WatchResponse)
	w.Header.Revision
	for t, v := range w.Events {
		revision := v.Kv.CreateRevision
		fmt.Printf("%s %q : %q\n", v.Type, v.Kv.Key, v.Kv.Value)
	}
	return nil, nil
}
