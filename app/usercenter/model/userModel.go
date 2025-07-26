package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel
		// 根据手机号查找用户
		FindOneByMobile(ctx context.Context, mobile string) (*User, error)
		// 根据昵称搜索用户（模糊查询）
		SearchByNickname(ctx context.Context, keyword string, page, limit int32) ([]*User, error)
		// 统计昵称搜索结果数量
		CountByNickname(ctx context.Context, keyword string) (int64, error)
	}

	customUserModel struct {
		*defaultUserModel
	}
)

// NewUserModel returns a model for the database table.
func NewUserModel(conn sqlx.SqlConn, c cache.CacheConf) UserModel {
	return &customUserModel{
		defaultUserModel: newUserModel(conn, c),
	}
}

// 根据手机号查找用户
func (m *customUserModel) FindOneByMobile(ctx context.Context, mobile string) (*User, error) {
	query := fmt.Sprintf("select %s from %s where `mobile` = ? and `del_state` = 0 limit 1", userRows, m.table)
	var resp User
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, mobile)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// 根据昵称搜索用户（模糊查询）
func (m *customUserModel) SearchByNickname(ctx context.Context, keyword string, page, limit int32) ([]*User, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := fmt.Sprintf("select %s from %s where `nickname` like ? and `del_state` = 0 order by `create_time` desc limit ?, ?", userRows, m.table)
	var resp []*User
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, "%"+keyword+"%", offset, limit)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// 统计昵称搜索结果数量
func (m *customUserModel) CountByNickname(ctx context.Context, keyword string) (int64, error) {
	query := "select count(*) from " + m.table + " where `nickname` like ? and `del_state` = 0"
	var count int64
	err := m.QueryRowNoCacheCtx(ctx, &count, query, "%"+keyword+"%")
	if err != nil {
		return 0, err
	}
	return count, nil
}
