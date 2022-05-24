package raft

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/store/pkg/config"
)

var raftId string
var bootstrap bool

//TODO:flags 配置
//var segmentSize int
//var noSync bool
//var segmentCacheSize int
//var binaryLogFormat bool

var dataDir string
var valueDir string

func InitFlags() {
	config.RegisterFlags(func(command *cobra.Command) {
		command.PersistentFlags().StringVar(&raftId, "RaftId", "", "Node id used by Raft(default hostname)")
		command.PersistentFlags().BoolVar(&bootstrap, "Bootstrap", false, "Whether to bootstrap the Raft cluster")

		//command.PersistentFlags().IntVar(&segmentSize, "Wal-SegmentSize", 100000, "wal SegmentSize")
		//command.PersistentFlags().BoolVar(&noSync, "Wal-NoSync", false, "wal NoSync")
		//command.PersistentFlags().IntVar(&segmentCacheSize, "Wal-SegmentCacheSize", 0, "wal SegmentCacheSize")
		//command.PersistentFlags().BoolVar(&binaryLogFormat, "Wal-BinaryLogFormat", false, "wal BinaryLogFormat")

		command.PersistentFlags().StringVar(&dataDir, "DataDir", "./data", "data dir")
		command.PersistentFlags().StringVar(&valueDir, "ValueDir", "./value", "value dir")
	})
}
