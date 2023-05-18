package store

import (
	"context"
	"github.com/dgraph-io/badger/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var KvStore *badger.DB

func Start(ctx context.Context) (err error) {
	dataDir := viper.GetString("DataDir")
	options := badger.DefaultOptions(dataDir)
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)
	//l.SetReportCaller(true)

	options.Logger = l
	KvStore, err = badger.Open(options)
	if err != nil {
		return err
	}
	go func() {
		select {
		case <-ctx.Done():
			KvStore.Close()
		}
	}()
	return nil
}
