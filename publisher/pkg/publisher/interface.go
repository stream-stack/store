package publisher

import (
	"context"
)

type SubscribeManager interface {
	LoadSnapshot(ctx context.Context, runner *SubscribeRunner) error
	Watch(ctx context.Context, runner *SubscribeRunner) error
	SaveSnapshot(ctx context.Context, runner *SubscribeRunner) error
}

type SubscribeManagerFactory func(ctx context.Context, storeAddress []string) (SubscribeManager, error)
