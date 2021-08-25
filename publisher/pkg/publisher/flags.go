package publisher

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/publisher/pkg/config"
)

var BackendTypeValue string
var BackendAddressValue []string

func InitFlags() {
	config.RegisterFlags(func(command *cobra.Command) {
		//command.PersistentFlags().StringVar(&StoreAddress, "StoreAddress", "localhost:5001", "store address")

		command.PersistentFlags().StringSliceVar(&BackendAddressValue, "BackendAddress", []string{"127.0.0.1:2379"}, "store address")
	})
}
