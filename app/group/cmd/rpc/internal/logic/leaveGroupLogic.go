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

type LeaveGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLeaveGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LeaveGroupLogic {
	return &LeaveGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 退出群组
func (l *LeaveGroupLogic) LeaveGroup(in *group.LeaveGroupReq) (*group.LeaveGroupResp, error) {
	// 参数验证
	if in.GroupId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid group id"), "groupId=%d", in.GroupId)
	}
	if in.UserId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", in.UserId)
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

	// 检查用户是否在群组中
	userMember, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, in.UserId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.NOT_FOUND, "user not in group"), "userId=%d, groupId=%d", in.UserId, in.GroupId)
		}
		return nil, errors.Wrapf(err, "check user member failed")
	}

	// 检查用户状态
	if userMember.Status != 1 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "user not active in group"), "userId=%d, status=%d", in.UserId, userMember.Status)
	}

	// 群主不能直接退群，必须先解散群或转让群主
	if userMember.Role == 3 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "owner cannot leave group, please dissolve or transfer ownership first"), "userId=%d", in.UserId)
	}

	// 使用事务处理退群
	err = l.svcCtx.ImGroupModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 更新用户成员状态为主动退出
		userMember.Status = 0 // 主动退出
		_, err = l.svcCtx.ImGroupMemberModel.Update(ctx, session, userMember)
		if err != nil {
			return errors.Wrapf(err, "update member status failed")
		}

		// 更新群组成员数量
		groupInfo.MemberCount--
		if groupInfo.MemberCount < 0 {
			groupInfo.MemberCount = 0
		}
		_, err = l.svcCtx.ImGroupModel.Update(ctx, session, groupInfo)
		if err != nil {
			return errors.Wrapf(err, "update group member count failed")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &group.LeaveGroupResp{
		Success: true,
	}, nil
}
