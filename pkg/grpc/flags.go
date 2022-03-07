package grpc

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/store/pkg/config"
	"time"
)

var applyLogTimeout time.Duration

func InitFlags() {
	config.RegisterFlags(func(command *cobra.Command) {
		command.PersistentFlags().DurationVar(&applyLogTimeout, "Grpc-ApplyLogTimeout", time.Second, "grpc apply log timeout second")
	})
}
