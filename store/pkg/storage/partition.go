package storage

import (
	"context"
	"fmt"
	"github.com/lafikl/consistent"
	"github.com/stream-stack/store/pkg/config"
)

type PartitionType string

type PartitionFunc func(streamName, streamId, eventId string, s []Storage) (Storage, error)

const HASHPartitionType = "HASH"

func HashPartitionFuncFactory(ctx context.Context, cfg *config.Config) (PartitionFunc, error) {
	c := consistent.New()
	for _, address := range cfg.StoreAddress {
		c.Add(address)
	}
	return func(streamName, streamId, eventId string, s []Storage) (Storage, error) {
		get, err := c.Get(formatKey(streamName, streamId, eventId))
		if err != nil {
			return nil, err
		}
		for _, storage := range s {
			if storage.GetAddress() == get {
				return storage, nil
			}
		}
		return nil, fmt.Errorf("[storage]consistent return %v,but storage address not Not included", get)
	}, nil
}

func formatKey(streamName, streamId, eventId string) string {
	return fmt.Sprintf("%s-%s-%s", streamName, streamId, eventId)
}
