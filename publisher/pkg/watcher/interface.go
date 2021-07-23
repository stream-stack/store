package watcher

import (
	"context"
	"github.com/stream-stack/monitor/pkg/config"
)

type Watcher interface {
	Watch(ctx context.Context) error
}

type Factory func(ctx context.Context, c *config.Config) ([]Watcher, error)
