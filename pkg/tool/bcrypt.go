package tool

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 加密方法
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败：%w", err)
	}
	return string(bytes), nil
}

// CheckPassword 检查密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err == nil
}
