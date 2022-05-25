package raft

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	"github.com/sirupsen/logrus"
	"io"
)

var fsmPrefix = []byte(`FSM:`)

type fsmImpl struct {
}

func (f *fsmImpl) Apply(log *raft.Log) interface{} {
	logrus.Debugf("fsm apply log{key:%v,index:%v,logType:%v}", string(log.Extensions), log.Index, log.Type)
	if log.Type != raft.LogCommand {
		logrus.Debugf("fsm apply log{key:%v,index:%v,logType:%v},log type not %v,skip", string(log.Extensions), log.Index, log.Type, raft.LogCommand)
		return nil
	}

	keyBytes := make([]byte, len(log.Extensions)+len(fsmPrefix))
	copy(keyBytes[:len(fsmPrefix)], fsmPrefix)
	copy(keyBytes[len(fsmPrefix):], log.Extensions)
	err := Store.conn.Update(func(txn *badger.Txn) error {
		return txn.Set(keyBytes, uint64ToBytes(log.Index))
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

func (f *fsmImpl) Snapshot() (raft.FSMSnapshot, error) {
	return &FSMSnapshotImpl{}, nil
}

func (f *fsmImpl) Restore(closer io.ReadCloser) error {
	defer closer.Close()
	return Store.conn.Load(closer, 1024)
}

type FSMSnapshotImpl struct {
}

func (F *FSMSnapshotImpl) Persist(sink raft.SnapshotSink) error {
	defer sink.Close()
	logrus.Debugf("begin snapshot persist")
	backup, err := Store.conn.Backup(sink, 0)
	if err != nil {
		logrus.Errorf("backup error:%+v", err)
		return err
	}
	logrus.Debugf("backup finish,backup:%+v", backup)

	return nil
}

func (F *FSMSnapshotImpl) Release() {
	logrus.Debugf("release snapshot")
}
