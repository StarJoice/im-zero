package model

import "fmt"

// GenerateConversationId 生成会话ID
func GenerateConversationId(userId, friendId int64) string {
	if userId < friendId {
		return fmt.Sprintf("%d_%d", userId, friendId)
	}
	return fmt.Sprintf("%d_%d", friendId, userId)
}