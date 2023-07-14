package subscriber

import (
	"github.com/spf13/cobra"
	"github.com/stream-stack/store/pkg/config"
	"time"
)

func InitFlags() {
	config.RegisterFlags(func(command *cobra.Command) {
		command.PersistentFlags().Duration("IndexOffsetStoreInterval", time.Second*10, "index store offset interval")
		command.PersistentFlags().Duration("SlotDigestOffsetStoreInterval", time.Minute*10, "slot digest store offset interval")
	})
}
