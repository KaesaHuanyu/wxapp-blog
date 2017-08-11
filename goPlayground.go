package main

import (
	"fmt"
	"time"
)

func _main() {
	fmt.Println(int(time.Now().UnixNano()))
}

func lengthOfLIS(arr []int) (num int) {
	if len(arr) == 0 || arr == nil {
		return 0
	}
	dp := []int{1}
	for i := 1; i < len(arr); i++ {
		min := false
		max := 0
		for j := 0; j < i; j++ {
			if arr[j] < arr[i] {
				min = true
				if dp[j] > max {
					max = dp[j]
				}
			}
		}
		if !min {
			dp = append(dp, 1)
		} else {
			dp = append(dp, max+1)
		}
	}
	for _, v := range dp {
		if v > num {
			num = v
		}
	}
	return
}
