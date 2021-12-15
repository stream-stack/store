package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

const Apply byte = 'A'
const Offset byte = 'O'

func AddApplyFlag(data []byte) []byte {
	tmp := make([]byte, len(data)+1)
	tmp[0] = Apply
	copy(tmp[1:], data)
	return tmp
}

func AddOffsetFlag(offset uint64) []byte {
	buffer := &bytes.Buffer{}
	buffer.WriteByte(Offset)
	binary.Write(buffer, binary.BigEndian, offset)
	return buffer.Bytes()
}

func FormatApplyMeta(streamName string, streamId string, eventId string) []byte {
	return []byte(fmt.Sprintf("%s/%s/%s", streamName, streamId, eventId))
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
