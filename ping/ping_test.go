package ping

import (
	"fmt"
	"testing"
)

func Test_ping(t *testing.T) {
	// 发起ping操作
	pingResult, _, err := Ping("192.168.221.155", "20.205.243.166", 10, 800, 10)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(pingResult)
}
