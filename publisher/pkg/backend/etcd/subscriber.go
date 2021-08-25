package etcd

import (
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"github.com/stream-stack/store/store/common/errdef"
	"github.com/stream-stack/store/store/common/formater"
	"github.com/stream-stack/store/store/publisher/pkg/publisher"
	"log"
	"reflect"
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

type etcdSnapshot struct {
	//快照创建位置
	StartPoint uint64 `json:"start_point"`
	//扩展数据
	ExtData []byte `json:"ext_data"`
}

func (e *etcdSnapshot) SetStartPoint(i publisher.StartPoint) {
	u := i.(uint64)
	e.StartPoint = u
}

func (e *etcdSnapshot) GetStartPoint() publisher.StartPoint {
	return e.StartPoint
}

func (e *etcdSnapshot) GetExtData() []byte {
	return e.ExtData
}

func (s *SubscribeManagerImpl) LoadSnapshot(ctx context.Context, streamName string, streamId string) (publisher.Snapshot, error) {
	var key string
	var err error
	var getResp *clientv3.GetResponse
	key = formater.FormatStreamInfo(streamName, streamId)
	getResp, err = s.etcdClient.Get(ctx, key, clientv3.WithSort(clientv3.SortByCreateRevision, clientv3.SortDescend),
		clientv3.WithPrefix(), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if getResp.Kvs == nil || len(getResp.Kvs) == 0 {
		return nil, errdef.ErrEventNotFound
	}
	if getResp.Kvs == nil {
	}
	value := getResp.Kvs[0]
	if value.CreateRevision == 0 {
		return nil, errdef.ErrEventNotFound
	}

	snapshot := &etcdSnapshot{}
	err = json.Unmarshal(value.Value, snapshot)
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (s *SubscribeManagerImpl) SaveSnapshot(ctx context.Context, streamName string, streamId string, data []byte) error {
	key := formater.FormatKey(streamName, streamId, time.Now().String())
	value := string(data)
	timeout, cancelFunc := context.WithTimeout(ctx, Timeout)
	defer cancelFunc()
	txn := s.etcdClient.Txn(timeout).If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).Then(clientv3.OpPut(key, value)).Else(clientv3.OpGet(key))
	commit, err := txn.Commit()
	if err != nil {
		return err
	}
	if commit.Succeeded {
		log.Printf("[etcd]key not exist, put key success,response:%+v", commit.OpResponse().Txn().Responses[0])
		return nil
	} else {
		log.Printf("[etcd]key exist, get key value response:%+v", commit.OpResponse().Txn().Responses[0])
		log.Printf("[etcd]data是否相等:%v \n", reflect.DeepEqual(commit.OpResponse().Txn().Responses[0].GetResponseRange().Kvs[0].Value, data))
		if !reflect.DeepEqual(commit.OpResponse().Txn().Responses[0].GetResponseRange().Kvs[0].Value, data) {
			return errdef.ErrEventExists
		}
	}
	return nil
}

func (s *SubscribeManagerImpl) Watch(ctx context.Context, point publisher.StartPoint, Key string, eu publisher.EventUnmarshaler) (<-chan publisher.SubscribeEvent, error) {
	i := point.(int64)
	watch := s.etcdClient.Watch(ctx, Key, clientv3.WithRev(i))
	c := make(chan publisher.SubscribeEvent)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(c)
			case response := <-watch:
				//反序列化错误处理
				events, _ := eu(Key, response)
				for _, event := range events {
					c <- event
				}
			}
		}
	}()
	return c, nil
}
