package etcd

import (
	"fmt"
	"github.com/stream-stack/publisher/pkg/publisher"
)

type etcdStartPoint map[string]uint64

func (s2 etcdStartPoint) Compare(s publisher.StartPoint) int {
	point, ok := s.(etcdStartPoint)
	if !ok {
		panic(fmt.Errorf("convert StartPoint to etcdStartPoint error,current type:%T", s))
	}
	sVal := point.minVal()
	s2Val := s2.minVal()
	if s2Val > sVal {
		return -1
	}
	if s2Val < sVal {
		return 1
	}
	return 0
}

func (s2 etcdStartPoint) minVal() uint64 {
	var minVal uint64
	for _, val := range s2 {
		if val < minVal {
			minVal = val
		}
	}
	return minVal
}

const SubscribeEventAdd = "ADD"
const SubscribeEventRemove = "REMOVE"

type SubscribeEvent struct {
	Operation string `json:"operation"`
	Name      string `json:"name"`
	Pattern   string `json:"pattern"`
	Url       string `json:"url"`
}
