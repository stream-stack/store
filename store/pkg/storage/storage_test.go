package storage

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestProxy_Get(t *testing.T) {
	storage1 := MockStorage{result: time.Now(), data: "test1"}
	storage2 := MockStorage{result: time.Now().Add(time.Hour), data: "test2"}
	proxy := Proxy{
		backend:       []Storage{&storage1, &storage2},
		partitionFunc: nil,
		address:       nil,
	}

	t.Run("getLast", func(t *testing.T) {
		got, err := proxy.Get(context.TODO(), "name1", "id1", LastEvent)
		if err != nil {
			t.Errorf("Get() got = %v, want %v", got, "test2")
		}
		if !reflect.DeepEqual(got, []byte("test2")) {
			t.Errorf("Get() got = %v, want %v", got, "test2")
		}
	})
	t.Run("getFirst", func(t *testing.T) {
		got, err := proxy.Get(context.TODO(), "name1", "id1", FirstEvent)
		if err != nil {
			t.Errorf("Get() got = %v, want %v", got, "test1")
		}
		if !reflect.DeepEqual(got, []byte("test1")) {
			t.Errorf("Get() got = %v, want %v", got, "test1")
		}
	})
}

type MockStorage struct {
	result time.Time
	data   string
}

func (m *MockStorage) Save(ctx context.Context, streamName, streamId, eventId string, data []byte) error {
	panic("implement me")
}

func (m *MockStorage) GetAddress() string {
	panic("implement me")
}

func (m *MockStorage) Get(ctx context.Context, streamName, streamId, eventId string) ([]byte, error) {
	marshal, err := json.Marshal(SaveRecord{
		Data:       []byte(m.data),
		CreateTime: m.result,
	})
	return marshal, err
}
