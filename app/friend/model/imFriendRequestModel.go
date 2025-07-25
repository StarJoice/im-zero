package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImFriendRequestModel = (*customImFriendRequestModel)(nil)

type (
	// ImFriendRequestModel is an interface to be customized, add more methods here,
	// and implement the added methods in customImFriendRequestModel.
	ImFriendRequestModel interface {
		imFriendRequestModel
		// 查找指定用户之间的好友请求
		FindOneByFromToUserId(ctx context.Context, fromUserId, toUserId int64) (*ImFriendRequest, error)
		// 获取用户收到的好友请求
		FindReceivedRequests(ctx context.Context, userId int64, page, limit int32) ([]*ImFriendRequest, error)
		// 获取用户发送的好友请求
		FindSentRequests(ctx context.Context, userId int64, page, limit int32) ([]*ImFriendRequest, error)
		// 统计用户待处理的好友请求数量
		CountPendingRequests(ctx context.Context, userId int64) (int64, error)
		// 获取用户所有待处理的好友请求
		FindPendingRequests(ctx context.Context, userId int64) ([]*ImFriendRequest, error)
		// 更新过期的好友请求
		UpdateExpiredRequests(ctx context.Context) error
	}

	customImFriendRequestModel struct {
		*defaultImFriendRequestModel
	}
)

// NewImFriendRequestModel returns a model for the database table.
func NewImFriendRequestModel(conn sqlx.SqlConn, c cache.CacheConf) ImFriendRequestModel {
	return &customImFriendRequestModel{
		defaultImFriendRequestModel: newImFriendRequestModel(conn, c),
	}
}

// 查找指定用户之间的好友请求
func (m *customImFriendRequestModel) FindOneByFromToUserId(ctx context.Context, fromUserId, toUserId int64) (*ImFriendRequest, error) {
	query := fmt.Sprintf("select %s from %s where `from_user_id` = ? and `to_user_id` = ? and `del_state` = 0 order by `create_time` desc limit 1", imFriendRequestRows, m.table)
	var resp ImFriendRequest
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, fromUserId, toUserId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// 获取用户收到的好友请求
func (m *customImFriendRequestModel) FindReceivedRequests(ctx context.Context, userId int64, page, limit int32) ([]*ImFriendRequest, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := fmt.Sprintf("select %s from %s where `to_user_id` = ? and `del_state` = 0 order by `create_time` desc limit ?, ?", imFriendRequestRows, m.table)
	var resp []*ImFriendRequest
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, userId, offset, limit)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// 获取用户发送的好友请求
func (m *customImFriendRequestModel) FindSentRequests(ctx context.Context, userId int64, page, limit int32) ([]*ImFriendRequest, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := fmt.Sprintf("select %s from %s where `from_user_id` = ? and `del_state` = 0 order by `create_time` desc limit ?, ?", imFriendRequestRows, m.table)
	var resp []*ImFriendRequest
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, userId, offset, limit)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// 统计用户待处理的好友请求数量
func (m *customImFriendRequestModel) CountPendingRequests(ctx context.Context, userId int64) (int64, error) {
	query := "select count(*) from " + m.table + " where `to_user_id` = ? and `status` = 0 and `del_state` = 0 and (`expire_time` is null or `expire_time` > now())"
	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, query, userId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// 获取用户所有待处理的好友请求
func (m *customImFriendRequestModel) FindPendingRequests(ctx context.Context, userId int64) ([]*ImFriendRequest, error) {
	query := fmt.Sprintf("select %s from %s where `to_user_id` = ? and `status` = 0 and `del_state` = 0 and (`expire_time` is null or `expire_time` > now()) order by `create_time` desc", imFriendRequestRows, m.table)
	var resp []*ImFriendRequest
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// 更新过期的好友请求
func (m *customImFriendRequestModel) UpdateExpiredRequests(ctx context.Context) error {
	query := "update " + m.table + " set `status` = 3, `update_time` = now() where `status` = 0 and `expire_time` is not null and `expire_time` <= now() and `del_state` = 0"
	_, err := m.ExecNoCacheCtx(ctx, query)
	return err
}
