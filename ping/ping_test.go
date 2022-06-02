package ping

import (
	"fmt"
	"testing"
)

func Test_ping(t *testing.T) {
	// 发起ping操作
	pingResult, _, err := Ping("0.0.0.0", "www.baidu.com", 32, 1000, 1000)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(pingResult)
}
