package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/stream-stack/store/store/common/formater"
	"github.com/stream-stack/store/store/common/proto"
	"github.com/stream-stack/store/store/publisher/pkg/publisher"
	"log"
	"time"
)

func NewSubscribeManagerFunc(ctx context.Context, storeAddress []string) (publisher.SubscribeManager, error) {
	var cfg clientv3.Config
	if len(Username) > 0 {
		cfg = clientv3.Config{
			Endpoints:   storeAddress,
			DialTimeout: Timeout,
			Context:     ctx,
			Username:    Username,
			Password:    Password,
		}
	} else {
		cfg = clientv3.Config{
			Endpoints:   storeAddress,
			DialTimeout: Timeout,
			Context:     ctx,
		}
	}
	cli, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	go func() {
		select {
		case <-ctx.Done():
			_ = cli.Close()
		}
	}()
	return &SubscribeManagerImpl{etcdClient: cli}, nil
}

type SubscribeManagerImpl struct {
	etcdClient *clientv3.Client
}

func (s *SubscribeManagerImpl) LoadSnapshot(ctx context.Context, runner *publisher.SubscribeRunner) error {
	var key string
	var err error
	var getResp *clientv3.GetResponse
	key = formater.FormatStreamInfo(runner.StreamName, runner.StreamId)
	getResp, err = s.etcdClient.Get(ctx, key, clientv3.WithSort(clientv3.SortByCreateRevision, clientv3.SortDescend),
		clientv3.WithPrefix(), clientv3.WithLimit(1))
	if err != nil {
		return err
	}
	if getResp.Kvs == nil || len(getResp.Kvs) == 0 {
		runner.StartPoint = int64(0)
		runner.ExtData = make(map[string]*proto.BaseSubscribe)
		return nil
	}
	value := getResp.Kvs[0]
	if value.CreateRevision == 0 {
		runner.StartPoint = int64(0)
		runner.ExtData = make(map[string]*proto.BaseSubscribe)
		return nil
	}

	snapshot := &etcdSnapshot{}
	err = json.Unmarshal(value.Value, snapshot)
	if err != nil {
		return err
	}
	runner.StartPoint = snapshot.StartPoint
	runner.ExtData = snapshot.Subscribes
	for _, subscribe := range snapshot.Subscribes {
		subRunner := publisher.NewSubscribeRunnerWithSubscribeOperation(runner, subscribe)
		err := subRunner.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SubscribeManagerImpl) SaveSnapshot(ctx context.Context, runner *publisher.SubscribeRunner) error {
	key := formater.FormatKey(runner.StreamName, runner.StreamId, time.Now().String())
	sp := runner.StartPoint.(int64)
	ed := runner.ExtData.(map[string]*proto.BaseSubscribe)
	e := &etcdSnapshot{
		StartPoint: sp,
		Subscribes: ed,
	}
	value, err := json.Marshal(e)
	if err != nil {
		return err
	}
	timeout, cancelFunc := context.WithTimeout(ctx, Timeout)
	defer cancelFunc()
	txn := s.etcdClient.Txn(timeout).If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).Then(clientv3.OpPut(key, string(value))).Else(clientv3.OpGet(key))
	commit, err := txn.Commit()
	if err != nil {
		return err
	}
	if commit.Succeeded {
		log.Printf("[etcd]key not exist, put key success,response:%+v", commit.OpResponse().Txn().Responses[0])
		return nil
	}
	return nil
}

func (s *SubscribeManagerImpl) Watch(ctx context.Context, runner *publisher.SubscribeRunner) error {
	i := runner.StartPoint.(int64)
	log.Printf("[etcd]watch key: %s,startPoint:%v", runner.WatchKey, i+1)
	watch := s.etcdClient.Watch(ctx, runner.WatchKey, clientv3.WithRev(i+1), clientv3.WithPrefix())
	for {
		select {
		case <-ctx.Done():
			return nil
		case w := <-watch:
			for _, v := range w.Events {
				fmt.Printf("%s %q : %q\n", v.Type, v.Kv.Key, v.Kv.Value)
				if v.Type == mvccpb.DELETE {
					continue
				}
				streamName, streamId, eventId, err := formater.SscanfKey(string(v.Kv.Key))
				if err != nil {
					log.Printf("[etcd]formater.SscanfKey key %s error:%v , ingore this key", string(v.Kv.Key), err)
					continue
				}
				revision := v.Kv.CreateRevision
				newEvent := cloudevents.NewEvent()
				newEvent.SetType(streamName)
				newEvent.SetSubject(streamName)
				newEvent.SetSource(streamId)
				newEvent.SetID(eventId)
				err = newEvent.SetData(cloudevents.ApplicationJSON, v.Kv.Value)
				if err != nil {
					return err
				}
				//TODO:如果一直出现错误,则需要死信队列
				err = runner.Action(ctx, newEvent, runner)
				if err != nil {
					return err
				}
				runner.StartPoint = revision
			}
		}
	}
}
