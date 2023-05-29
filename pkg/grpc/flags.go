package grpc

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/store/pkg/config"
	"time"
)

func InitFlags() {
	config.RegisterFlags(func(command *cobra.Command) {
		command.PersistentFlags().String("Port", "50051", "grpc port")
		command.PersistentFlags().Duration("OffsetStoreInterval", time.Second*10, "store offset interval")
	})
}
