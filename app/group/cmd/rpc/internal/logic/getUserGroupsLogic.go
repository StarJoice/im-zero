package logic

import (
	"context"
	"im-zero/pkg/tool"

	"im-zero/app/group/cmd/rpc/group"
	"im-zero/app/group/cmd/rpc/internal/svc"
	"im-zero/pkg/xerrs"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserGroupsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserGroupsLogic {
	return &GetUserGroupsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用户的群组列表
func (l *GetUserGroupsLogic) GetUserGroups(in *group.GetUserGroupsReq) (*group.GetUserGroupsResp, error) {
	// 参数验证
	if in.UserId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", in.UserId)
	}

	// 设置默认值
	page := in.Page
	if page <= 0 {
		page = 1
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// 先获取用户的群成员关系
	memberBuilder := l.svcCtx.ImGroupMemberModel.SelectBuilder().
		Where(squirrel.Eq{"user_id": in.UserId}).
		Where(squirrel.Eq{"status": 1}) // 只查询正常状态的成员关系

	members, memberTotal, err := l.svcCtx.ImGroupMemberModel.FindPageListByPageWithTotal(
		l.ctx,
		memberBuilder,
		int64(page),
		int64(limit),
		"join_time DESC",
	)
	if err != nil {
		return nil, errors.Wrapf(err, "find user group members failed")
	}

	if len(members) == 0 {
		return &group.GetUserGroupsResp{
			Groups: []*group.GroupInfo{},
			Total:  0,
		}, nil
	}

	// 提取群组ID列表
	var groupIds []int64
	for _, member := range members {
		groupIds = append(groupIds, member.GroupId)
	}

	// 批量获取群组信息
	groupBuilder := l.svcCtx.ImGroupModel.SelectBuilder().
		Where(squirrel.Eq{"id": groupIds}).
		Where(squirrel.Eq{"status": 1}) // 只查询正常状态的群组

	groups, err := l.svcCtx.ImGroupModel.FindAll(l.ctx, groupBuilder, "update_time DESC")
	if err != nil {
		return nil, errors.Wrapf(err, "find groups failed")
	}

	// 转换为响应格式
	var groupInfos []*group.GroupInfo
	for _, g := range groups {
		groupInfo := &group.GroupInfo{
			Id:                g.Id,
			Name:              g.Name,
			Avatar:            g.Avatar.String,
			Description:       g.Description.String,
			Notice:            g.Notice.String,
			OwnerId:           g.OwnerId,
			MemberCount:       int32(g.MemberCount),
			MaxMembers:        int32(g.MaxMembers),
			Status:            int32(g.Status),
			IsPrivate:         tool.Int64ToBool(g.IsPrivate),
			JoinApproval:      tool.Int64ToBool(g.JoinApproval),
			AllowInvite:       tool.Int64ToBool(g.AllowInvite),
			AllowMemberModify: tool.Int64ToBool(g.AllowMemberModify),
			CreateTime:        g.CreateTime.Unix(),
			UpdateTime:        g.UpdateTime.Unix(),
		}
		groupInfos = append(groupInfos, groupInfo)
	}

	return &group.GetUserGroupsResp{
		Groups: groupInfos,
		Total:  int32(memberTotal),
	}, nil
}
