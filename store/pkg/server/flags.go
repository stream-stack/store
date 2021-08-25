package server

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/store/store/common/config"
)

var GrpcPort string

func InitFlags() {
	config.RegisterFlags(func(command *cobra.Command) {
		command.PersistentFlags().StringVar(&GrpcPort, "GrpcPort", "5001", "grpc port")
	})
}
