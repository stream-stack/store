package config

import "github.com/spf13/cobra"

var DataDir string
var Address string

func InitFlags() {
	RegisterFlags(func(command *cobra.Command) {
		command.PersistentFlags().StringVar(&Address, "Address", "localhost:50051", "TCP host+port for this node")
		command.PersistentFlags().StringVar(&DataDir, "DataDir", "data", "data dir")
	})
}
