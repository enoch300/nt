package mtr

import (
	"fmt"
	"testing"
)

func Test_mtr(t *testing.T) {
	// 发起mtr操作
	mtrResult, _, err := Mtr("0.0.0.0", "221.131.143.69", 32, 1, 800)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(mtrResult)
}
