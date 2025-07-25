package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImUserBlacklistModel = (*customImUserBlacklistModel)(nil)

type (
	// ImUserBlacklistModel is an interface to be customized, add more methods here,
	// and implement the added methods in customImUserBlacklistModel.
	ImUserBlacklistModel interface {
		imUserBlacklistModel
		// 查找指定用户之间的拉黑关系
		FindOneByUserIdBlockedUserId(ctx context.Context, userId, blockedUserId int64) (*ImUserBlacklist, error)
		// 获取用户的黑名单列表
		FindBlacklistByUserId(ctx context.Context, userId int64) ([]*ImUserBlacklist, error)
		// 检查用户是否被拉黑
		CheckBlocked(ctx context.Context, userId, targetUserId int64) (bool, error)
		// 批量检查用户是否被拉黑
		BatchCheckBlocked(ctx context.Context, userId int64, targetUserIds []int64) (map[int64]bool, error)
		// 获取拉黑用户的数量
		CountBlacklistByUserId(ctx context.Context, userId int64) (int64, error)
		// 双向检查拉黑状态
		CheckMutualBlocked(ctx context.Context, userId1, userId2 int64) (user1BlockedUser2, user2BlockedUser1 bool, err error)
	}

	customImUserBlacklistModel struct {
		*defaultImUserBlacklistModel
	}
)

// NewImUserBlacklistModel returns a model for the database table.
func NewImUserBlacklistModel(conn sqlx.SqlConn, c cache.CacheConf) ImUserBlacklistModel {
	return &customImUserBlacklistModel{
		defaultImUserBlacklistModel: newImUserBlacklistModel(conn, c),
	}
}

// 查找指定用户之间的拉黑关系
func (m *customImUserBlacklistModel) FindOneByUserIdBlockedUserId(ctx context.Context, userId, blockedUserId int64) (*ImUserBlacklist, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `blocked_user_id` = ? and `del_state` = 0 limit 1", imUserBlacklistRows, m.table)
	var resp ImUserBlacklist
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, userId, blockedUserId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// 获取用户的黑名单列表
func (m *customImUserBlacklistModel) FindBlacklistByUserId(ctx context.Context, userId int64) ([]*ImUserBlacklist, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `del_state` = 0 order by `create_time` desc", imUserBlacklistRows, m.table)
	var resp []*ImUserBlacklist
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// 检查用户是否被拉黑
func (m *customImUserBlacklistModel) CheckBlocked(ctx context.Context, userId, targetUserId int64) (bool, error) {
	query := "select count(*) from " + m.table + " where `user_id` = ? and `blocked_user_id` = ? and `del_state` = 0"
	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, query, userId, targetUserId)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 批量检查用户是否被拉黑
func (m *customImUserBlacklistModel) BatchCheckBlocked(ctx context.Context, userId int64, targetUserIds []int64) (map[int64]bool, error) {
	if len(targetUserIds) == 0 {
		return make(map[int64]bool), nil
	}

	// 构建IN查询
	query := fmt.Sprintf("select `blocked_user_id` from %s where `user_id` = ? and `blocked_user_id` in (", m.table)
	args := make([]interface{}, len(targetUserIds)+1)
	args[0] = userId
	for i, id := range targetUserIds {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i+1] = id
	}
	query += ") and `del_state` = 0"

	var results []struct {
		BlockedUserId int64 `db:"blocked_user_id"`
	}
	err := m.QueryRowsNoCacheCtx(ctx, &results, query, args...)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]bool)
	// 先设置所有为false
	for _, id := range targetUserIds {
		result[id] = false
	}
	// 再设置被拉黑的为true
	for _, item := range results {
		result[item.BlockedUserId] = true
	}

	return result, nil
}

// 获取拉黑用户的数量
func (m *customImUserBlacklistModel) CountBlacklistByUserId(ctx context.Context, userId int64) (int64, error) {
	query := "select count(*) from " + m.table + " where `user_id` = ? and `del_state` = 0"
	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, query, userId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// 双向检查拉黑状态
func (m *customImUserBlacklistModel) CheckMutualBlocked(ctx context.Context, userId1, userId2 int64) (user1BlockedUser2, user2BlockedUser1 bool, err error) {
	// 检查userId1是否拉黑了userId2
	user1BlockedUser2, err = m.CheckBlocked(ctx, userId1, userId2)
	if err != nil {
		return false, false, err
	}

	// 检查userId2是否拉黑了userId1
	user2BlockedUser1, err = m.CheckBlocked(ctx, userId2, userId1)
	if err != nil {
		return false, false, err
	}

	return user1BlockedUser2, user2BlockedUser1, nil
}
