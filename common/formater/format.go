package formater

import (
	"fmt"
	"github.com/stream-stack/store/store/common/vars"
)

func FormatStreamInfo(name string, id string) string {
	return fmt.Sprintf("/%s/%s/%s/", vars.StorePrefix, name, id)
}

func FormatKey(streamName, streamId, eventId string) string {
	return fmt.Sprintf("/%s/%s/%s/%s", vars.StorePrefix, streamName, streamId, eventId)
}
