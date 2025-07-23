package tool

import "regexp"

// ValidateMobile 验证手机号格式
// 支持格式:
// 1. 中国大陆手机号 (11位数字, 以1开头)
// 2. 带+86前缀的手机号
// 3. 带空格或短横线的分隔格式
func ValidateMobile(mobile string) bool {
	// 去除空格和短横线
	re := regexp.MustCompile(`[-\s]`)
	mobile = re.ReplaceAllString(mobile, "")

	// 匹配规则:
	// 1. 可选的+86前缀
	// 2. 11位数字, 以1开头
	pattern := `^(?:\+?86)?1[3-9]\d{9}$`
	return regexp.MustCompile(pattern).MatchString(mobile)
}
