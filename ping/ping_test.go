package ping

import (
	"fmt"
	"testing"
)

func Test_ping(t *testing.T) {
	// 发起ping操作
	pingResult, _, err := Ping("0.0.0.0", "183.238.66.11", 15, 800, 10)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(pingResult)
}
