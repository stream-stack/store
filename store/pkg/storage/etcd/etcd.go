package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/stream-stack/store/store/common/errdef"
	"github.com/stream-stack/store/store/common/formater"
	"github.com/stream-stack/store/store/common/vars"
	"github.com/stream-stack/store/store/pkg/storage"
	"log"
	"reflect"
)

const BackendType = "ETCD"

type etcdConnect struct {
	cli          *clientv3.Client
	addressSlice []string
}

func (e *etcdConnect) Get(ctx context.Context, streamName, streamId, eventId string) ([]byte, error) {
	var key string
	var err error
	var getResp *clientv3.GetResponse
	switch eventId {
	case vars.FirstEvent:
		key = formater.FormatStreamInfo(streamName, streamId)
		getResp, err = e.cli.Get(ctx, key, clientv3.WithSort(clientv3.SortByCreateRevision, clientv3.SortAscend),
			clientv3.WithPrefix(), clientv3.WithLimit(1))
	case vars.LastEvent:
		key = formater.FormatStreamInfo(streamName, streamId)
		getResp, err = e.cli.Get(ctx, key, clientv3.WithSort(clientv3.SortByCreateRevision, clientv3.SortDescend),
			clientv3.WithPrefix(), clientv3.WithLimit(1))
	default:
		getResp, err = e.cli.Get(ctx, formater.FormatKey(streamName, streamId, eventId))
	}

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
	return value.Value, nil
}

func (e *etcdConnect) Save(ctx context.Context, streamName, streamId, eventId string, data []byte) error {
	key := formater.FormatKey(streamName, streamId, eventId)
	value := string(data)
	timeout, cancelFunc := context.WithTimeout(ctx, Timeout)
	defer cancelFunc()
	txn := e.cli.Txn(timeout).If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).Then(clientv3.OpPut(key, value)).Else(clientv3.OpGet(key))
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

func (e *etcdConnect) GetCurrentStartPoint(ctx context.Context) (interface{}, error) {
	put, err := e.cli.Put(ctx, "SetStartPointTemp", "")
	if err != nil {
		return nil, err
	}
	return put.Header.Revision, nil
}

func (e *etcdConnect) start(ctx context.Context) error {
	log.Printf("[etcd]start etcd,params[address:%v,username:%v,password:%v]", e.addressSlice, Username, Password)
	var cfg clientv3.Config
	if len(Username) > 0 {
		cfg = clientv3.Config{
			Endpoints:   e.addressSlice,
			DialTimeout: Timeout,
			Context:     ctx,
			Username:    Username,
			Password:    Password,
		}
	} else {
		cfg = clientv3.Config{
			Endpoints:   e.addressSlice,
			DialTimeout: Timeout,
			Context:     ctx,
		}
	}
	var err error
	e.cli, err = clientv3.New(cfg)
	if err != nil {
		return err
	}
	go func() {
		select {
		case <-ctx.Done():
			_ = e.cli.Close()
		}
	}()
	return nil
}

func NewStorageFunc(ctx context.Context, addressSlice []string) (storage.Storage, error) {
	connect := &etcdConnect{addressSlice: addressSlice}
	if err := connect.start(ctx); err != nil {
		return nil, err
	}
	return connect, nil
}
