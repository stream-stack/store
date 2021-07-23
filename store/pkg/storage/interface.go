package storage

import (
	"context"
	"errors"
	"github.com/stream-stack/store/pkg/config"
)

type Storage interface {
	Save(ctx context.Context, streamName, streamId, eventId string, data []byte) error
	GetAddress() string
	Get(ctx context.Context, streamName, streamId, eventId string) ([]byte, error)
}

type Factory func(ctx context.Context, c *config.Config) ([]Storage, error)

const FirstEvent = "FIRST"
const LastEvent = "LAST"

var ErrEventNotFound = errors.New("event not found")
var ErrEventExists = errors.New("event exists")
