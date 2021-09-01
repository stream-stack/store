package formater

import (
	"fmt"
	"testing"
)

func TestSscanfKey(t *testing.T) {
	key, id, i, err := SscanfKey("/stream/system/subscribe/2021-09-01 11:58:18.2509027 +0800 CST m=+1536.008895201")
	if err != nil {
		panic(err)
	}
	fmt.Println(key, id, i)
}
