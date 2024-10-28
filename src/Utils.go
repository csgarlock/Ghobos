package main

func clampInt32(x int32, min int32, max int32) int32 {
	if x > max {
		return max
	} else if x < min {
		return min
	} else {
		return x
	}
}
