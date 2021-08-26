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

func SscanfKey(key string) (streamName, streamId, eventId string, err error) {
	var count int
	count, err = fmt.Sscanf(key, "/"+vars.StorePrefix+"/%s/%s/%s", &streamName, &streamId, &eventId)
	if err != nil {
		return
	}
	if count != 3 {
		err = fmt.Errorf("parse key error,Unable to parse complete data")
	}
	return
}
