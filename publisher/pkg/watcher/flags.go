package watcher

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/publisher/pkg/config"
)

func InitFlags() {
	config.RegisterFlags(func(command *cobra.Command) {

	})
}
