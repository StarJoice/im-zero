package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImGroupRequestModel = (*customImGroupRequestModel)(nil)

type (
	// ImGroupRequestModel is an interface to be customized, add more methods here,
	// and implement the added methods in customImGroupRequestModel.
	ImGroupRequestModel interface {
		imGroupRequestModel
	}

	customImGroupRequestModel struct {
		*defaultImGroupRequestModel
	}
)

// NewImGroupRequestModel returns a model for the database table.
func NewImGroupRequestModel(conn sqlx.SqlConn, c cache.CacheConf) ImGroupRequestModel {
	return &customImGroupRequestModel{
		defaultImGroupRequestModel: newImGroupRequestModel(conn, c),
	}
}
