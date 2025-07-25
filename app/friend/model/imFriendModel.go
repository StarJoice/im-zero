package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImFriendModel = (*customImFriendModel)(nil)

type (
	// ImFriendModel is an interface to be customized, add more methods here,
	// and implement the added methods in customImFriendModel.
	ImFriendModel interface {
		imFriendModel
		// 查找用户的好友关系
		FindOneByUserIdFriendId(ctx context.Context, userId, friendId int64) (*ImFriend, error)
		// 获取用户的好友列表
		FindFriendsByUserId(ctx context.Context, userId int64) ([]*ImFriend, error)
		// 获取用户的好友数量
		CountFriendsByUserId(ctx context.Context, userId int64) (int64, error)
		// 检查好友关系是否存在
		CheckFriendship(ctx context.Context, userId, friendId int64) (bool, error)
		// 批量检查好友关系
		BatchCheckFriendship(ctx context.Context, userId int64, friendIds []int64) (map[int64]bool, error)
		// 根据备注搜索好友
		SearchFriendsByRemark(ctx context.Context, userId int64, keyword string) ([]*ImFriend, error)
	}

	customImFriendModel struct {
		*defaultImFriendModel
	}
)

// NewImFriendModel returns a model for the database table.
func NewImFriendModel(conn sqlx.SqlConn, c cache.CacheConf) ImFriendModel {
	return &customImFriendModel{
		defaultImFriendModel: newImFriendModel(conn, c),
	}
}

// 查找用户的好友关系
func (m *customImFriendModel) FindOneByUserIdFriendId(ctx context.Context, userId, friendId int64) (*ImFriend, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `friend_id` = ? and `del_state` = 0 limit 1", imFriendRows, m.table)
	var resp ImFriend
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, userId, friendId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// 获取用户的好友列表
func (m *customImFriendModel) FindFriendsByUserId(ctx context.Context, userId int64) ([]*ImFriend, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `del_state` = 0 and `status` = 1 order by `is_top` desc, `create_time` desc", imFriendRows, m.table)
	var resp []*ImFriend
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// 获取用户的好友数量
func (m *customImFriendModel) CountFriendsByUserId(ctx context.Context, userId int64) (int64, error) {
	query := "select count(*) from " + m.table + " where `user_id` = ? and `del_state` = 0 and `status` = 1"
	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, query, userId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// 检查好友关系是否存在
func (m *customImFriendModel) CheckFriendship(ctx context.Context, userId, friendId int64) (bool, error) {
	query := "select count(*) from " + m.table + " where `user_id` = ? and `friend_id` = ? and `del_state` = 0 and `status` = 1"
	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, query, userId, friendId)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 批量检查好友关系
func (m *customImFriendModel) BatchCheckFriendship(ctx context.Context, userId int64, friendIds []int64) (map[int64]bool, error) {
	if len(friendIds) == 0 {
		return make(map[int64]bool), nil
	}

	// 构建IN查询
	query := fmt.Sprintf("select `friend_id` from %s where `user_id` = ? and `friend_id` in (", m.table)
	args := make([]interface{}, len(friendIds)+1)
	args[0] = userId
	for i, id := range friendIds {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i+1] = id
	}
	query += ") and `del_state` = 0 and `status` = 1"

	var results []struct {
		FriendId int64 `db:"friend_id"`
	}
	err := m.QueryRowsNoCacheCtx(ctx, &results, query, args...)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]bool)
	// 先设置所有为false
	for _, id := range friendIds {
		result[id] = false
	}
	// 再设置存在的为true
	for _, item := range results {
		result[item.FriendId] = true
	}

	return result, nil
}

// 根据备注搜索好友
func (m *customImFriendModel) SearchFriendsByRemark(ctx context.Context, userId int64, keyword string) ([]*ImFriend, error) {
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `remark` like ? and `del_state` = 0 and `status` = 1 order by `create_time` desc", imFriendRows, m.table)
	var resp []*ImFriend
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, userId, "%"+keyword+"%")
	if err != nil {
		return nil, err
	}
	return resp, nil
}
