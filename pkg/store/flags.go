package store

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/store/pkg/config"
)

func InitFlags() {
	config.RegisterFlags(func(c *cobra.Command) {
		c.PersistentFlags().String("DataDir", "./data", "data dir")
	})
}
