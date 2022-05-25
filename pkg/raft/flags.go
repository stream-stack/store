package raft

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/store/pkg/config"
	"time"
)

//raft config flags
var raftId string
var bootstrap bool

//snapshot config
var snapshotRetain int
var snapshotDataDir string

//badger config
var badgerSync bool
var badgerValueLogGC bool
var badgerGCInterval time.Duration
var badgerMandatoryGCInterval time.Duration
var badgerGCThreshold int64

//dir config
var baseDir string
var dataDir string
var valueDir string

func InitFlags() {
	config.RegisterFlags(func(command *cobra.Command) {
		command.PersistentFlags().StringVar(&raftId, "RaftId", "", "Node id used by Raft(default hostname)")
		command.PersistentFlags().BoolVar(&bootstrap, "Bootstrap", false, "Whether to bootstrap the Raft cluster")

		command.PersistentFlags().StringVar(&snapshotDataDir, "snapshotDataDir", "snapshot", "snapshot data directory")
		command.PersistentFlags().IntVar(&snapshotRetain, "snapshotRetain", 10, "snapshot retain count")

		command.PersistentFlags().StringVar(&baseDir, "BaseDir", "./data", "base data directory")
		command.PersistentFlags().StringVar(&dataDir, "DataDir", "data", "data dir")
		command.PersistentFlags().StringVar(&valueDir, "ValueDir", "value", "value dir")

		command.PersistentFlags().BoolVar(&badgerSync, "badgerSync", true, "badger Sync")
		command.PersistentFlags().BoolVar(&badgerValueLogGC, "badgerValueLogGC", false, "badger ValueLogGC")
		command.PersistentFlags().DurationVar(&badgerGCInterval, "badgerGCInterval", time.Minute, "badger GCInterval")
		command.PersistentFlags().Int64Var(&badgerGCThreshold, "badgerGCThreshold", 1<<30, "badger GCThreshold")
		command.PersistentFlags().DurationVar(&badgerMandatoryGCInterval, "badgerMandatoryGCInterval", time.Minute*10, "badger MandatoryGCInterval")
	})
}
