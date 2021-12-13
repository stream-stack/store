package wal

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/store/pkg/config"
)

func InitFlags() {
	config.RegisterFlags(func(command *cobra.Command) {
	})
}
