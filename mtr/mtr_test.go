package mtr

import (
	"fmt"
	"testing"
)

func Test_mtr(t *testing.T) {
	// 发起mtr操作
	mtrResult, _, err := Mtr("192.168.221.159", "8.8.8.8", 30, 5, 800)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(mtrResult)
}
