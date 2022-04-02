package protocol

import (
	"encoding/binary"
	"fmt"
	"strings"
)

const Apply byte = 'A'
const KeyValue byte = 'K'

func AddApplyFlag(data []byte) []byte {
	tmp := make([]byte, len(data)+1)
	tmp[0] = Apply
	copy(tmp[1:], data)
	return tmp
}

func AddKeyValueFlag(data []byte) []byte {
	tmp := make([]byte, len(data)+1)
	tmp[0] = KeyValue
	copy(tmp[1:], data)
	return tmp
}

func FormatApplyMeta(streamName string, streamId string, eventId uint64) []byte {
	return []byte(fmt.Sprintf("%s/%s/%d", streamName, streamId, eventId))
}

func ParseMeta(meta []byte) ([]string, error) {
	s := string(meta)
	split := strings.Split(s, "/")
	if len(split) != 3 {
		return nil, fmt.Errorf("parse meta %s error", s)
	}
	return split, nil
}

// BytesToUint64 Converts bytes to an integer
func BytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// Uint64ToBytes Converts a uint to a byte slice
func Uint64ToBytes(u uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, u)
	return buf
}
