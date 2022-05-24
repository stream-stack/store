package raft

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	"github.com/sirupsen/logrus"
	"io"
)

type fsmImpl struct {
}

func (f *fsmImpl) Apply(log *raft.Log) interface{} {
	logrus.Debugf("fsm apply log{key:%v,index:%v,logType:%v}", string(log.Extensions), log.Index, log.Type)
	if log.Type != raft.LogCommand {
		logrus.Debugf("fsm apply log{key:%v,index:%v,logType:%v},log type not %v,skip", string(log.Extensions), log.Index, log.Type, raft.LogCommand)
		return nil
	}

	err := Store.conn.Update(func(txn *badger.Txn) error {
		return txn.Set(log.Extensions, uint64ToBytes(log.Index))
	})
	if err != nil {
		return err
	}

	f2 := func(watchers []*watcher) {
		logrus.Debugf("notify watchers size:%d", len(watchers))
		notCleanWatcher := make([]*watcher, 0, len(watchers))
		for _, w := range watchers {
			if w.filter != nil {
				filter, err := w.filter(log)
				if err != nil {
					logrus.Errorf("execute filter error:%+v,log: %+v", err, log)
					notCleanWatcher = append(notCleanWatcher, w)
					continue
				}
				if !filter {
					notCleanWatcher = append(notCleanWatcher, w)
					continue
				}
			}
			close(w.c)
			logrus.Debugf("close subscribe %s notify channel finish", w.id)
		}
		updateWaiters(notCleanWatcher)
	}
	watcherOpCh <- f2

	return err
}

//TODO:实现构建快照
func (f *fsmImpl) Snapshot() (raft.FSMSnapshot, error) {
	return &FSMSnapshotImpl{}, nil
}

func (f *fsmImpl) Restore(closer io.ReadCloser) error {
	return closer.Close()
}

type FSMSnapshotImpl struct {
}

func (F *FSMSnapshotImpl) Persist(sink raft.SnapshotSink) error {
	return nil
}

func (F *FSMSnapshotImpl) Release() {

}
