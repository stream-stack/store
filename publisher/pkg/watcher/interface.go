package watcher

import (
	"context"
)

type Watcher interface {
	Watch(ctx context.Context) error
}

type Factory func(ctx context.Context, addressSlice []string) ([]Watcher, error)
