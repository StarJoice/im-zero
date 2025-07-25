package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImGroupMessageModel = (*customImGroupMessageModel)(nil)

type (
	// ImGroupMessageModel is an interface to be customized, add more methods here,
	// and implement the added methods in customImGroupMessageModel.
	ImGroupMessageModel interface {
		imGroupMessageModel
	}

	customImGroupMessageModel struct {
		*defaultImGroupMessageModel
	}
)

// NewImGroupMessageModel returns a model for the database table.
func NewImGroupMessageModel(conn sqlx.SqlConn, c cache.CacheConf) ImGroupMessageModel {
	return &customImGroupMessageModel{
		defaultImGroupMessageModel: newImGroupMessageModel(conn, c),
	}
}
