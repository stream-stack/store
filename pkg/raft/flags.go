package raft

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/store/pkg/config"
)

var RaftId string
var Bootstrap bool

func InitFlags() {
	config.RegisterFlags(func(command *cobra.Command) {
		command.PersistentFlags().StringVar(&RaftId, "RaftId", "", "Node id used by Raft")
		command.PersistentFlags().BoolVar(&Bootstrap, "Bootstrap", false, "Whether to bootstrap the Raft cluster")
	})
}
