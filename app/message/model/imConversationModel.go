package model

import (
	"context"
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImConversationModel = (*customImConversationModel)(nil)

type (
	// ImConversationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customImConversationModel.
	ImConversationModel interface {
		imConversationModel
		// 获取用户的所有会话列表
		FindConversationsByUserId(ctx context.Context, userId int64, page, pageSize int64) ([]*ImConversation, int64, error)
		// 更新会话的最后消息信息
		UpdateLastMessage(ctx context.Context, session sqlx.Session, conversationId string, messageId int64, content string) error
		// 增加未读计数
		IncrementUnreadCount(ctx context.Context, session sqlx.Session, conversationId string, userId int64) error
		// 清零未读计数
		ClearUnreadCount(ctx context.Context, session sqlx.Session, conversationId string, userId int64) error
		// 查找或创建会话
		FindOrCreateConversation(ctx context.Context, session sqlx.Session, userId, friendId int64) (*ImConversation, error)
	}

	customImConversationModel struct {
		*defaultImConversationModel
	}
)

// NewImConversationModel returns a model for the database table.
func NewImConversationModel(conn sqlx.SqlConn, c cache.CacheConf) ImConversationModel {
	return &customImConversationModel{
		defaultImConversationModel: newImConversationModel(conn, c),
	}
}

// FindConversationsByUserId 获取用户的所有会话列表
func (m *customImConversationModel) FindConversationsByUserId(ctx context.Context, userId int64, page, pageSize int64) ([]*ImConversation, int64, error) {
	query := m.SelectBuilder().Where("user_id = ?", userId)
	return m.FindPageListByPageWithTotal(ctx, query, page, pageSize, "update_time DESC")
}

// UpdateLastMessage 更新会话的最后消息信息
func (m *customImConversationModel) UpdateLastMessage(ctx context.Context, session sqlx.Session, conversationId string, messageId int64, content string) error {
	query := "UPDATE " + m.table + " SET last_message_id = ?, last_message_content = ?, last_message_time = NOW() WHERE conversation_id = ?"

	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		if session != nil {
			return session.ExecCtx(ctx, query, messageId, content, conversationId)
		}
		return conn.ExecCtx(ctx, query, messageId, content, conversationId)
	})
	return err
}

// IncrementUnreadCount 增加未读计数
func (m *customImConversationModel) IncrementUnreadCount(ctx context.Context, session sqlx.Session, conversationId string, userId int64) error {
	query := "UPDATE " + m.table + " SET unread_count = unread_count + 1 WHERE conversation_id = ? AND user_id = ?"

	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		if session != nil {
			return session.ExecCtx(ctx, query, conversationId, userId)
		}
		return conn.ExecCtx(ctx, query, conversationId, userId)
	})
	return err
}

// ClearUnreadCount 清零未读计数
func (m *customImConversationModel) ClearUnreadCount(ctx context.Context, session sqlx.Session, conversationId string, userId int64) error {
	query := "UPDATE " + m.table + " SET unread_count = 0 WHERE conversation_id = ? AND user_id = ?"

	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		if session != nil {
			return session.ExecCtx(ctx, query, conversationId, userId)
		}
		return conn.ExecCtx(ctx, query, conversationId, userId)
	})
	return err
}

// FindOrCreateConversation 查找或创建会话
func (m *customImConversationModel) FindOrCreateConversation(ctx context.Context, session sqlx.Session, userId, friendId int64) (*ImConversation, error) {
	// 每个用户有自己的会话记录，使用用户特定的会话ID
	userConversationId := GenerateUserConversationId(userId, friendId)

	// 先尝试通过会话ID查找现有会话
	conversation, err := m.FindOneByConversationId(ctx, userConversationId)
	if err == nil {
		return conversation, nil
	}

	if err != ErrNotFound {
		return nil, err
	}

	// 如果不存在，创建新会话
	newConversation := &ImConversation{
		ConversationId:   userConversationId,
		UserId:           userId,
		FriendId:         friendId,
		ConversationType: 1, // 1-单聊
		UnreadCount:      0,
		IsTop:            0,
		IsMute:           0,
	}

	_, err = m.Insert(ctx, session, newConversation)
	if err != nil {
		// 如果是重复键错误，再次尝试查找现有会话（可能是并发创建）
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			conversation, findErr := m.FindOneByConversationId(ctx, userConversationId)
			if findErr == nil {
				return conversation, nil
			}
		}
		return nil, err
	}

	return newConversation, nil
}
