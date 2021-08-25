package etcd

type etcdStartPoint uint64

const SubscribeEventAdd = "ADD"
const SubscribeEventRemove = "REMOVE"

type SubscribeEvent struct {
	Operation string `json:"operation"`
	Name      string `json:"name"`
	Pattern   string `json:"pattern"`
	Url       string `json:"url"`
}
