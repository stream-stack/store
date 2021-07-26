package storage

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/store/pkg/config"
)

var PartitionValue string
var StoreTypeValue string
var StoreAddressValue []string

func InitFlags() {
	config.RegisterFlags(func(command *cobra.Command) {
		command.PersistentFlags().StringVar(&PartitionValue, "PartitionType", HASHPartitionType, "partition type")
	})
}
