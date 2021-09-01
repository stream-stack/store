package formater

import (
	"fmt"
	"github.com/stream-stack/store/store/common/vars"
	"regexp"
)

func FormatStreamInfo(name string, id string) string {
	return fmt.Sprintf("/%s/%s/%s/", vars.StorePrefix, name, id)
}

func FormatKey(streamName, streamId, eventId string) string {
	return fmt.Sprintf("%s/%s/%s/%s", vars.StorePrefix, streamName, streamId, eventId)
}

var compile *regexp.Regexp
var regStr = vars.StorePrefix + "/(.+)/(.+)/(.+)"

func init() {
	compile = regexp.MustCompile(regStr)
}

func SscanfKey(key string) (streamName, streamId, eventId string, err error) {
	matchString := compile.MatchString(key)
	if !matchString {
		err = fmt.Errorf("%s not match %s", key, regStr)
		return
	}
	submatch := compile.FindStringSubmatch(key)
	if len(submatch) != 4 {
		err = fmt.Errorf("parse key error,Unable to parse complete data")
		return
	}
	streamName = submatch[1]
	streamId = submatch[2]
	eventId = submatch[3]
	return
}
