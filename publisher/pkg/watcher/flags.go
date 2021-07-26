package watcher

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/monitor/pkg/config"
)

var StoreTypeValue string
var StoreAddressValue []string

func InitFlags() {
	config.RegisterFlags(func(command *cobra.Command) {
	})
}
