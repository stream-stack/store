package storage

import (
	"context"
	"fmt"
	"github.com/lafikl/consistent"
)

type PartitionType string

type PartitionFunc func(streamName, streamId, eventId string, s []Storage) (Storage, error)

const HASHPartitionType = "HASH"

func HashPartitionFuncFactory(ctx context.Context, addressSlice []string) (PartitionFunc, error) {
	c := consistent.New()
	for _, address := range addressSlice {
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
		return nil, fmt.Errorf("[storage]consistent return %v,but storage addressSlice not Not included", get)
	}, nil
}

func formatKey(streamName, streamId, eventId string) string {
	return fmt.Sprintf("%s-%s-%s", streamName, streamId, eventId)
}
