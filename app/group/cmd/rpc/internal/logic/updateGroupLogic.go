package logic

import (
	"context"
	"database/sql"

	"im-zero/app/group/cmd/rpc/group"
	"im-zero/app/group/cmd/rpc/internal/svc"
	"im-zero/app/group/model"
	"im-zero/pkg/tool"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UpdateGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGroupLogic {
	return &UpdateGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新群组信息
func (l *UpdateGroupLogic) UpdateGroup(in *group.UpdateGroupReq) (*group.UpdateGroupResp, error) {
	// 参数验证
	if in.GroupId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid group id"), "groupId=%d", in.GroupId)
	}
	if in.OperatorId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid operator id"), "operatorId=%d", in.OperatorId)
	}

	// 获取群组信息
	groupInfo, err := l.svcCtx.ImGroupModel.FindOne(l.ctx, in.GroupId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.NOT_FOUND, "group not found"), "groupId=%d", in.GroupId)
		}
		return nil, errors.Wrapf(err, "find group failed")
	}

	// 检查群组状态
	if groupInfo.Status != 1 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "group is not active"), "groupId=%d, status=%d", in.GroupId, groupInfo.Status)
	}

	// 检查操作者权限
	operatorMember, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, in.OperatorId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "operator not in group"), "operatorId=%d, groupId=%d", in.OperatorId, in.GroupId)
		}
		return nil, errors.Wrapf(err, "check operator member failed")
	}

	// 检查操作者状态
	if operatorMember.Status != 1 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "operator not active in group"), "operatorId=%d, status=%d", in.OperatorId, operatorMember.Status)
	}

	// 检查是否有修改权限
	canModify := false
	if operatorMember.Role == 3 { // 群主可以修改所有信息
		canModify = true
	} else if operatorMember.Role == 2 && groupInfo.AllowMemberModify == 1 { // 管理员在允许的情况下可以修改
		canModify = true
	} else if operatorMember.Role == 1 && groupInfo.AllowMemberModify == 1 { // 普通成员在允许的情况下可以修改
		canModify = true
	}

	if !canModify {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.GROUP_PERMISSION_DENIED, "no permission to modify group info"), "operatorId=%d, role=%d", in.OperatorId, operatorMember.Role)
	}

	// 使用事务处理群组信息更新
	err = l.svcCtx.ImGroupModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 参数长度验证
		if len(in.Name) > 100 {
			return errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "group name too long"), "name length=%d", len(in.Name))
		}
		if len(in.Description) > 500 {
			return errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "group description too long"), "description length=%d", len(in.Description))
		}

		// 更新群组信息
		if len(in.Name) > 0 {
			groupInfo.Name = in.Name
		}
		if len(in.Avatar) > 0 {
			groupInfo.Avatar = sql.NullString{String: in.Avatar, Valid: true}
		}
		if len(in.Description) > 0 {
			groupInfo.Description = sql.NullString{String: in.Description, Valid: true}
		}
		if len(in.Notice) > 0 {
			groupInfo.Notice = sql.NullString{String: in.Notice, Valid: true}
		}

		// 只有群主和管理员可以修改这些设置
		if operatorMember.Role >= 2 {
			groupInfo.JoinApproval = tool.BoolToInt64(in.JoinApproval)
			groupInfo.AllowInvite = tool.BoolToInt64(in.AllowInvite)
		}

		// 只有群主可以修改这个设置
		if operatorMember.Role == 3 {
			groupInfo.AllowMemberModify = tool.BoolToInt64(in.AllowMemberModify)
		}

		_, err = l.svcCtx.ImGroupModel.Update(ctx, session, groupInfo)
		if err != nil {
			return errors.Wrapf(err, "update group info failed")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &group.UpdateGroupResp{
		Success: true,
	}, nil
}
