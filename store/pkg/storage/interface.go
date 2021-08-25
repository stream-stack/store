package storage

import (
	"context"
)

type Storage interface {
	Save(ctx context.Context, streamName, streamId, eventId string, data []byte) error
	GetAddress() string
	Get(ctx context.Context, streamName, streamId, eventId string) ([]byte, error)
}

type Factory func(ctx context.Context, addressSlice []string) (Storage, error)
