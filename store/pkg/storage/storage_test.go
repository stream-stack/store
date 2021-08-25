package storage

import (
	"context"
	"encoding/json"
	"github.com/stream-stack/store/store/common/vars"
	"reflect"
	"testing"
	"time"
)

func TestProxy_Get(t *testing.T) {
	storage1 := MockStorage{result: time.Now(), data: "test1"}

	t.Run("getLast", func(t *testing.T) {
		got, err := storage1.Get(context.TODO(), "name1", "id1", vars.LastEvent)
		if err != nil {
			t.Errorf("Get() got = %v, want %v", got, "test2")
		}
		if !reflect.DeepEqual(got, []byte("test1")) {
			t.Errorf("Get() got = %v, want %v", got, "test2")
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
		CreateTime: time.Now(),
	})
	return marshal, err
}
