package logic

import (
	"context"

	"im-zero/app/group/cmd/rpc/group"
	"im-zero/app/group/cmd/rpc/internal/svc"
	"im-zero/app/group/model"
	"im-zero/pkg/xerrs"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupMembersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupMembersLogic {
	return &GetGroupMembersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取群成员列表
func (l *GetGroupMembersLogic) GetGroupMembers(in *group.GetGroupMembersReq) (*group.GetGroupMembersResp, error) {
	// 参数验证
	if in.GroupId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid group id"), "groupId=%d", in.GroupId)
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

	// 检查群组是否存在
	groupInfo, err := l.svcCtx.ImGroupModel.FindOne(l.ctx, in.GroupId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.NOT_FOUND, "group not found"), "groupId=%d", in.GroupId)
		}
		return nil, errors.Wrapf(err, "find group failed")
	}

	// 检查群组状态
	if groupInfo.Status != 1 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.NOT_FOUND, "group is not active"), "groupId=%d, status=%d", in.GroupId, groupInfo.Status)
	}

	// 构建查询条件
	rowBuilder := l.svcCtx.ImGroupMemberModel.SelectBuilder().
		Where(squirrel.Eq{"group_id": in.GroupId}).
		Where(squirrel.Eq{"status": 1}) // 只查询正常状态的成员

	// 如果指定了角色筛选
	if in.Role > 0 {
		rowBuilder = rowBuilder.Where(squirrel.Eq{"role": in.Role})
	}

	// 获取成员列表和总数
	members, total, err := l.svcCtx.ImGroupMemberModel.FindPageListByPageWithTotal(
		l.ctx, 
		rowBuilder, 
		int64(page), 
		int64(limit), 
		"role ASC, join_time ASC", // 按角色和入群时间排序
	)
	if err != nil {
		return nil, errors.Wrapf(err, "find group members failed")
	}

	// 转换为响应格式
	var memberInfos []*group.GroupMemberInfo
	for _, member := range members {
		memberInfo := &group.GroupMemberInfo{
			UserId:      member.UserId,
			GroupId:     member.GroupId,
			Nickname:    member.Nickname.String,
			Avatar:      member.Avatar.String,
			Role:        int32(member.Role),
			Status:      int32(member.Status),
			JoinTime:    member.JoinTime.Unix(),
		}

		// 设置禁言结束时间
		if member.MuteEndTime.Valid {
			memberInfo.MuteEndTime = member.MuteEndTime.Time.Unix()
		}

		memberInfos = append(memberInfos, memberInfo)
	}

	return &group.GetGroupMembersResp{
		Members: memberInfos,
		Total:   int32(total),
	}, nil
}
