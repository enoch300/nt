package mtr

import (
	"fmt"
	"testing"
)

func Test_mtr(t *testing.T) {
	// 发起mtr操作
	mtrResult, _, err := Mtr("192.168.221.155", "20.205.243.166", 30, 5, 800)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(mtrResult)
}
