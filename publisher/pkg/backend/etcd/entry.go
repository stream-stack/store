package etcd

import (
	"github.com/stream-stack/store/store/publisher/pkg/publisher"
)

type etcdSnapshot struct {
	//快照创建位置
	StartPoint uint64 `json:"start_point"`
	//扩展数据
	Subscribes map[string]*publisher.BaseSubscribe `json:"subscribes"`
}
