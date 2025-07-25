package logic

import (
	"context"

	"im-zero/app/group/cmd/rpc/group"
	"im-zero/app/group/cmd/rpc/internal/svc"
	"im-zero/app/group/model"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SetMemberRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetMemberRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetMemberRoleLogic {
	return &SetMemberRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置成员角色
func (l *SetMemberRoleLogic) SetMemberRole(in *group.SetMemberRoleReq) (*group.SetMemberRoleResp, error) {
	// 参数验证
	if in.GroupId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid group id"), "groupId=%d", in.GroupId)
	}
	if in.OperatorId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid operator id"), "operatorId=%d", in.OperatorId)
	}
	if in.UserId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", in.UserId)
	}
	if in.Role < 1 || in.Role > 3 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid role"), "role=%d", in.Role)
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

	// 只有群主可以设置角色
	if operatorMember.Role != 3 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "only owner can set member role"), "operatorId=%d, role=%d", in.OperatorId, operatorMember.Role)
	}

	// 不能操作自己
	if in.UserId == in.OperatorId {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "cannot set own role"), "userId=%d", in.UserId)
	}

	// 不能设置群主角色
	if in.Role == 3 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "cannot set owner role"), "role=%d", in.Role)
	}

	// 使用事务处理角色更新
	err = l.svcCtx.ImGroupModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 检查目标用户是否在群中
		targetMember, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(ctx, in.GroupId, in.UserId)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.NOT_FOUND, "target user not in group"), "userId=%d, groupId=%d", in.UserId, in.GroupId)
			}
			return errors.Wrapf(err, "check target member failed")
		}

		// 检查目标用户状态
		if targetMember.Status != 1 {
			return errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "target user not active in group"), "userId=%d, status=%d", in.UserId, targetMember.Status)
		}

		// 不能修改群主的角色
		if targetMember.Role == 3 {
			return errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "cannot modify owner role"), "userId=%d", in.UserId)
		}

		// 更新角色
		targetMember.Role = int64(in.Role)
		_, err = l.svcCtx.ImGroupMemberModel.Update(ctx, session, targetMember)
		if err != nil {
			return errors.Wrapf(err, "update member role failed")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &group.SetMemberRoleResp{
		Success: true,
	}, nil
}
