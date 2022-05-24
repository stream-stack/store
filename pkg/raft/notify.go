package raft

import (
	"context"
	"github.com/hashicorp/raft"
	"github.com/sirupsen/logrus"
)

type watcher struct {
	c      chan struct{}
	filter func(log *raft.Log) (bool, error)
	id     string
}

var watchers = make([]*watcher, 0)
var watcherOpCh = make(chan func(waiters []*watcher))

func updateWaiters(w []*watcher) {
	logrus.Debugf("update Waiters to size:%d", len(w))
	watchers = w
}

func StartNotify(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case f := <-watcherOpCh:
			f(watchers)
		}
	}
}

func Wait(filter func(log *raft.Log) (bool, error), id string) <-chan struct{} {
	c := make(chan struct{})
	w := &watcher{
		id:     id,
		c:      c,
		filter: filter,
	}

	watcherOpCh <- func(waiters []*watcher) {
		waiters = append(waiters, w)
		updateWaiters(waiters)
	}
	return c
}
