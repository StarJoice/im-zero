package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImGroupMemberModel = (*customImGroupMemberModel)(nil)

type (
	// ImGroupMemberModel is an interface to be customized, add more methods here,
	// and implement the added methods in customImGroupMemberModel.
	ImGroupMemberModel interface {
		imGroupMemberModel
		// 检查用户在群组中的权限
		CheckMemberPermission(ctx context.Context, groupId, userId int64) (*MemberPermission, error)
		// 获取群组的管理员列表
		FindAdminsByGroupId(ctx context.Context, groupId int64) ([]*ImGroupMember, error)
		// 统计群组各角色成员数量
		CountMembersByRole(ctx context.Context, groupId int64) (map[int64]int64, error)
	}

	customImGroupMemberModel struct {
		*defaultImGroupMemberModel
	}

	// 成员权限信息
	MemberPermission struct {
		IsMember    bool  `json:"is_member"`
		Role        int64 `json:"role"`         // 1-普通成员 2-管理员 3-群主
		Status      int64 `json:"status"`       // 0-已退出 1-正常 2-被踢出
		IsMuted     bool  `json:"is_muted"`     // 是否被禁言
		MuteEndTime *sql.NullTime `json:"mute_end_time"`
	}
)

// NewImGroupMemberModel returns a model for the database table.
func NewImGroupMemberModel(conn sqlx.SqlConn, c cache.CacheConf) ImGroupMemberModel {
	return &customImGroupMemberModel{
		defaultImGroupMemberModel: newImGroupMemberModel(conn, c),
	}
}

// 检查用户在群组中的权限
func (m *customImGroupMemberModel) CheckMemberPermission(ctx context.Context, groupId, userId int64) (*MemberPermission, error) {
	member, err := m.FindOneByGroupIdUserId(ctx, groupId, userId)
	if err != nil {
		if err == ErrNotFound {
			return &MemberPermission{IsMember: false}, nil
		}
		return nil, err
	}

	isMuted := false
	if member.MuteEndTime.Valid && member.MuteEndTime.Time.After(time.Now()) {
		isMuted = true
	}

	return &MemberPermission{
		IsMember:    member.Status == 1,
		Role:        member.Role,
		Status:      member.Status,
		IsMuted:     isMuted,
		MuteEndTime: &member.MuteEndTime,
	}, nil
}

// 获取群组的管理员列表（包括群主）
func (m *customImGroupMemberModel) FindAdminsByGroupId(ctx context.Context, groupId int64) ([]*ImGroupMember, error) {
	query := fmt.Sprintf("select %s from %s where `group_id` = ? and `role` in (2, 3) and `status` = 1 order by `role` desc, `join_time` asc", imGroupMemberRows, m.table)
	var resp []*ImGroupMember
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, groupId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// 统计群组各角色成员数量
func (m *customImGroupMemberModel) CountMembersByRole(ctx context.Context, groupId int64) (map[int64]int64, error) {
	query := "select `role`, count(*) as count from " + m.table + " where `group_id` = ? and `status` = 1 group by `role`"
	
	var results []struct {
		Role  int64 `db:"role"`
		Count int64 `db:"count"`
	}
	
	err := m.QueryRowsNoCacheCtx(ctx, &results, query, groupId)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]int64)
	for _, item := range results {
		result[item.Role] = item.Count
	}
	
	return result, nil
}
