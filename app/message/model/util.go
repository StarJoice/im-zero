package model

import "fmt"

// GenerateConversationId 生成会话ID（用于唯一标识一对用户间的会话）
func GenerateConversationId(userId, friendId int64) string {
	if userId < friendId {
		return fmt.Sprintf("%d_%d", userId, friendId)
	}
	return fmt.Sprintf("%d_%d", friendId, userId)
}

// GenerateUserConversationId 生成用户会话ID（每个用户有自己的会话记录）
func GenerateUserConversationId(userId, friendId int64) string {
	return fmt.Sprintf("%d_%d", userId, friendId)
}
