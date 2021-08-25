package formater

import (
	"fmt"
	"github.com/stream-stack/common/vars"
)

func formatStreamInfo(name string, id string) string {
	return fmt.Sprintf("/%s/%s/%s/", vars.StorePrefix, name, id)
}

func formatKey(streamName, streamId, eventId string) string {
	return fmt.Sprintf("/%s/%s/%s/%s", vars.StorePrefix, streamName, streamId, eventId)
}
