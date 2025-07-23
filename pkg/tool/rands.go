package tool

import (
	"crypto/rand"
	"math/big"
)

const (
	KC_RAND_KIND_NUM   = 0 // 纯数字
	KC_RAND_KIND_LOWER = 1 // 小写字母
	KC_RAND_KIND_UPPER = 2 // 大写字母
	KC_RAND_KIND_ALL   = 3 // 数字、大小写字母
)

// Krand 生成随机字符串
// size: 需要生成的字符串长度
// kind: 0=纯数字, 1=小写字母, 2=大写字母, 3=大小写数字混合
func Krand(size int, kind int) string {
	kinds := [][]int{
		{10, 48}, // 数字: 0-9
		{26, 97}, // 小写字母: a-z
		{26, 65}, // 大写字母: A-Z
	}

	buf := make([]byte, size)

	// 确定字符类型范围
	useAll := kind < 0 || kind > 3
	if useAll {
		kind = 3 // 默认使用混合模式
	}

	for i := range buf {
		var charType int

		// 如果是混合模式，随机选择字符类型
		if kind == 3 {
			n, _ := rand.Int(rand.Reader, big.NewInt(3))
			charType = int(n.Int64())
		} else {
			charType = kind
		}

		// 生成随机字符
		scope := kinds[charType][0]
		base := kinds[charType][1]

		n, _ := rand.Int(rand.Reader, big.NewInt(int64(scope)))
		buf[i] = byte(base + int(n.Int64()))
	}

	return string(buf)
}
