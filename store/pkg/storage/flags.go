package storage

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/store/store/common/config"
)

var BackendTypeValue string
var BackendAddressValue []string

func InitFlags() {
	config.RegisterFlags(func(command *cobra.Command) {
		command.PersistentFlags().StringSliceVar(&BackendAddressValue, "BackendAddress", []string{"127.0.0.1:2379"}, "store address")
	})
}
