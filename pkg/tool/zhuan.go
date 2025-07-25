package tool

// Int64ToBool 辅助函数
func Int64ToBool(i int64) bool {
	return i == 1
}

func BoolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
