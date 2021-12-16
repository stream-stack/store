package wal

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/store/pkg/config"
)

var segmentSize int
var noSync bool
var segmentCacheSize int
var binaryLogFormat bool

func InitFlags() {
	config.RegisterFlags(func(command *cobra.Command) {
		command.PersistentFlags().IntVar(&segmentSize, "Wal-SegmentSize", 100000, "wal SegmentSize")
		command.PersistentFlags().BoolVar(&noSync, "Wal-NoSync", false, "wal NoSync")
		command.PersistentFlags().IntVar(&segmentCacheSize, "Wal-SegmentCacheSize", 0, "wal SegmentCacheSize")
		command.PersistentFlags().BoolVar(&binaryLogFormat, "Wal-BinaryLogFormat", false, "wal BinaryLogFormat")
	})
}
