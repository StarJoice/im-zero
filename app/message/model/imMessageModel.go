package model

import (
	"context"
	"database/sql"
	"strings"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImMessageModel = (*customImMessageModel)(nil)

type (
	// ImMessageModel is an interface to be customized, add more methods here,
	// and implement the added methods in customImMessageModel.
	ImMessageModel interface {
		imMessageModel
		// 按会话ID分页查询消息记录（倒序）
		FindMessagesByConversationId(ctx context.Context, conversationId string, lastMessageId int64, limit int) ([]*ImMessage, error)
		// 获取用户未读消息数
		CountUnreadMessages(ctx context.Context, userId int64, conversationId string) (int64, error)
		// 批量更新消息状态
		UpdateMessagesStatus(ctx context.Context, session sqlx.Session, messageIds []int64, status int64) error
		// 更新单个消息状态
		UpdateMessageStatus(ctx context.Context, messageId int64, status int64) error
		// 获取会话中的最新消息
		FindLatestMessageByConversationId(ctx context.Context, conversationId string) (*ImMessage, error)
	}

	customImMessageModel struct {
		*defaultImMessageModel
	}
)

// NewImMessageModel returns a model for the database table.
func NewImMessageModel(conn sqlx.SqlConn, c cache.CacheConf) ImMessageModel {
	return &customImMessageModel{
		defaultImMessageModel: newImMessageModel(conn, c),
	}
}

// FindMessagesByConversationId 按会话ID分页查询消息记录（倒序）
func (m *customImMessageModel) FindMessagesByConversationId(ctx context.Context, conversationId string, lastMessageId int64, limit int) ([]*ImMessage, error) {
	query := m.SelectBuilder().Where("conversation_id = ?", conversationId)
	
	if lastMessageId > 0 {
		query = query.Where("id < ?", lastMessageId)
	}
	
	return m.FindAll(ctx, query.OrderBy("id DESC").Limit(uint64(limit)), "")
}

// CountUnreadMessages 获取用户未读消息数
func (m *customImMessageModel) CountUnreadMessages(ctx context.Context, userId int64, conversationId string) (int64, error) {
	query := m.SelectBuilder().Where("to_user_id = ? AND conversation_id = ? AND status < ?", userId, conversationId, 3) // status < 3 表示未读
	return m.FindCount(ctx, query, "id")
}

// UpdateMessagesStatus 批量更新消息状态
func (m *customImMessageModel) UpdateMessagesStatus(ctx context.Context, session sqlx.Session, messageIds []int64, status int64) error {
	if len(messageIds) == 0 {
		return nil
	}
	
	query := "UPDATE " + m.table + " SET status = ? WHERE id IN (" + 
		strings.Repeat("?,", len(messageIds)-1) + "?)"
	
	args := make([]interface{}, 0, len(messageIds)+1)
	args = append(args, status)
	for _, id := range messageIds {
		args = append(args, id)
	}
	
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		if session != nil {
			return session.ExecCtx(ctx, query, args...)
		}
		return conn.ExecCtx(ctx, query, args...)
	})
	return err
}

// UpdateMessageStatus 更新单个消息状态
func (m *customImMessageModel) UpdateMessageStatus(ctx context.Context, messageId int64, status int64) error {
	query := "UPDATE " + m.table + " SET status = ? WHERE id = ?"
	
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, query, status, messageId)
	})
	return err
}

// FindLatestMessageByConversationId 获取会话中的最新消息
func (m *customImMessageModel) FindLatestMessageByConversationId(ctx context.Context, conversationId string) (*ImMessage, error) {
	query := m.SelectBuilder().Where("conversation_id = ?", conversationId).OrderBy("id DESC").Limit(1)
	
	messages, err := m.FindAll(ctx, query, "")
	if err != nil {
		return nil, err
	}
	
	if len(messages) == 0 {
		return nil, ErrNotFound
	}
	
	return messages[0], nil
}
