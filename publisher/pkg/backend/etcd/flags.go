package etcd

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/spf13/cobra"
	"github.com/stream-stack/store/store/common/config"
	"github.com/stream-stack/store/store/common/formater"
	"github.com/stream-stack/store/store/publisher/pkg/publisher"
	"time"
)

const BackendType = "ETCD"

func InitFlags() {
	publisher.Register(BackendType, NewSubscribeManagerFunc)
	publisher.RegisterExtDataUnmarshal(BackendType, NoneExtDataUnmarshaler)
	publisher.RegisterEventUnmarshal(BackendType, cloudEventUnmarshaler)

	config.RegisterFlags(func(command *cobra.Command) {
		command.PersistentFlags().StringVar(&publisher.BackendTypeValue, "BackendType", BackendType, "store type")

		command.PersistentFlags().StringVar(&Username, "EtcdUsername", "", "etcd Username")
		command.PersistentFlags().StringVar(&Password, "EtcdPassword", "", "etcd Password")
		command.PersistentFlags().DurationVar(&Timeout, "EtcdTimeout", time.Second*5, "etcd connect Timeout")
	})
}

var Username string
var Password string
var Timeout time.Duration

func NoneExtDataUnmarshaler(data []byte) (interface{}, error) {
	return nil, nil
}

func cloudEventUnmarshaler(key string, event interface{}) ([]publisher.SubscribeEvent, error) {
	w := event.(clientv3.WatchResponse)

	events := make([]publisher.SubscribeEvent, 0)
	for _, v := range w.Events {
		fmt.Printf("%s %q : %q\n", v.Type, v.Kv.Key, v.Kv.Value)
		if v.Type == mvccpb.DELETE {
			continue
		}
		streamName, streamId, eventId, err := formater.SscanfKey(key)
		if err != nil {
			return nil, err
		}
		revision := v.Kv.CreateRevision
		e := &etcdEvent{
			StartPoint: uint64(revision),
		}
		newEvent := cloudevents.NewEvent()
		newEvent.SetType(streamName)
		newEvent.SetSubject(streamName)
		newEvent.SetSource(streamId)
		newEvent.SetID(eventId)
		_ = newEvent.SetData(cloudevents.ApplicationJSON, v.Kv.Value)
		e.Event = newEvent
		events = append(events, e)
	}
	return events, nil
}
