package etcd

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/stream-stack/store/pkg/storage"
	"log"
	"reflect"
)

const BackendType = "ETCD"
const StorePrefix = "stream"

type etcdConnect struct {
	address string
	cli     *clientv3.Client
}

func (e *etcdConnect) Get(ctx context.Context, streamName, streamId, eventId string) ([]byte, error) {
	var key string
	var err error
	var getResp *clientv3.GetResponse
	switch eventId {
	case storage.FirstEvent:
		key = formatStreamInfo(streamName, streamId)
		getResp, err = e.cli.Get(ctx, key, clientv3.WithSort(clientv3.SortByCreateRevision, clientv3.SortAscend),
			clientv3.WithPrefix(), clientv3.WithLimit(1))
	case storage.LastEvent:
		key = formatStreamInfo(streamName, streamId)
		getResp, err = e.cli.Get(ctx, key, clientv3.WithSort(clientv3.SortByCreateRevision, clientv3.SortDescend),
			clientv3.WithPrefix(), clientv3.WithLimit(1))
	default:
		getResp, err = e.cli.Get(ctx, formatKey(streamName, streamId, eventId))
	}

	if err != nil {
		return nil, err
	}
	if getResp.Kvs == nil || len(getResp.Kvs) == 0 {
		return nil, storage.ErrEventNotFound
	}
	if getResp.Kvs == nil {
	}
	value := getResp.Kvs[0]
	if value.CreateRevision == 0 {
		return nil, storage.ErrEventNotFound
	}
	return value.Value, nil
}

func formatStreamInfo(name string, id string) string {
	return fmt.Sprintf("/%s/%s/%s/", StorePrefix, name, id)
}

func (e *etcdConnect) Save(ctx context.Context, streamName, streamId, eventId string, data []byte) error {
	key := formatKey(streamName, streamId, eventId)
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
			return storage.ErrEventExists
		}
	}
	return nil
}

func formatKey(streamName, streamId, eventId string) string {
	return fmt.Sprintf("/%s/%s/%s/%s", StorePrefix, streamName, streamId, eventId)
}

func (e *etcdConnect) GetAddress() string {
	return e.address
}

func (e *etcdConnect) start(ctx context.Context, address string) error {
	log.Printf("[etcd]start etcd,params[address:%v,username:%v,password:%v]", address, Username, Password)
	var cfg clientv3.Config
	if len(Username) > 0 {
		cfg = clientv3.Config{
			Endpoints:   []string{address},
			DialTimeout: Timeout,
			Context:     ctx,
			Username:    Username,
			Password:    Password,
		}
	} else {
		cfg = clientv3.Config{
			Endpoints:   []string{address},
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

func NewStorageFunc(ctx context.Context, addressSlice []string) ([]storage.Storage, error) {
	storages := make([]storage.Storage, len(addressSlice))
	for i, address := range addressSlice {
		connect := &etcdConnect{address: address}
		if err := connect.start(ctx, address); err != nil {
			return nil, err
		}
		storages[i] = connect
	}
	return storages, nil
}
