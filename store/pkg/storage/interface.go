package storage

import (
	"context"
)

type Storage interface {
	Save(ctx context.Context, streamName, streamId, eventId string, data []byte) error
	GetCurrentStartPoint(ctx context.Context) (interface{}, error)
	Get(ctx context.Context, streamName, streamId, eventId string) ([]byte, error)
}

type Factory func(ctx context.Context, addressSlice []string) (Storage, error)
