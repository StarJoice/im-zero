package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImGroupModel = (*customImGroupModel)(nil)

type (
	// ImGroupModel is an interface to be customized, add more methods here,
	// and implement the added methods in customImGroupModel.
	ImGroupModel interface {
		imGroupModel
		// 获取群组的下一个消息序号
		GetNextMessageSeq(ctx context.Context, session sqlx.Session, groupId int64) (int64, error)
		// 批量获取群组信息
		FindByIds(ctx context.Context, ids []int64) ([]*ImGroup, error)
	}

	customImGroupModel struct {
		*defaultImGroupModel
	}
)

// NewImGroupModel returns a model for the database table.
func NewImGroupModel(conn sqlx.SqlConn, c cache.CacheConf) ImGroupModel {
	return &customImGroupModel{
		defaultImGroupModel: newImGroupModel(conn, c),
	}
}

// 获取群组的下一个消息序号
func (m *customImGroupModel) GetNextMessageSeq(ctx context.Context, session sqlx.Session, groupId int64) (int64, error) {
	// 使用群组表的version字段作为消息序号生成器
	query := "UPDATE " + m.table + " SET version = version + 1 WHERE id = ?"
	_, err := session.ExecCtx(ctx, query, groupId)
	if err != nil {
		return 0, err
	}

	// 获取更新后的version值
	var seq int64
	selectQuery := "SELECT version FROM " + m.table + " WHERE id = ?"
	err = session.QueryRowCtx(ctx, &seq, selectQuery, groupId)
	if err != nil {
		return 0, err
	}

	return seq, nil
}

// 批量获取群组信息
func (m *customImGroupModel) FindByIds(ctx context.Context, ids []int64) ([]*ImGroup, error) {
	if len(ids) == 0 {
		return []*ImGroup{}, nil
	}

	// 构建IN查询
	query := fmt.Sprintf("select %s from %s where `id` in (", imGroupRows, m.table)
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ") and `del_state` = 0 order by `id`"

	var resp []*ImGroup
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, args...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
