package storage

import (
	"context"
	"errors"
)

type Storage interface {
	Save(ctx context.Context, streamName, streamId, eventId string, data []byte) error
	GetAddress() string
	Get(ctx context.Context, streamName, streamId, eventId string) ([]byte, error)
}

type Factory func(ctx context.Context, addressSlice []string) (Storage, error)

const FirstEvent = "FIRST"
const LastEvent = "LAST"

var ErrEventNotFound = errors.New("event not found")
var ErrEventExists = errors.New("event exists")
