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
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DissolveGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDissolveGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DissolveGroupLogic {
	return &DissolveGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 解散群组
func (l *DissolveGroupLogic) DissolveGroup(in *group.DissolveGroupReq) (*group.DissolveGroupResp, error) {
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

	// 只有群主可以解散群组
	if operatorMember.Role != 3 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "only owner can dissolve group"), "operatorId=%d, role=%d", in.OperatorId, operatorMember.Role)
	}

	// 使用事务处理群组解散
	err = l.svcCtx.ImGroupModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 更新群组状态为已解散
		groupInfo.Status = 2 // 解散状态
		_, err = l.svcCtx.ImGroupModel.Update(ctx, session, groupInfo)
		if err != nil {
			return errors.Wrapf(err, "update group status failed")
		}

		// 将所有成员状态更新为0（已退出）
		memberBuilder := l.svcCtx.ImGroupMemberModel.SelectBuilder().
			Where(squirrel.Eq{"group_id": in.GroupId}).
			Where(squirrel.Eq{"status": 1}) // 只更新正常状态的成员

		members, err := l.svcCtx.ImGroupMemberModel.FindAll(ctx, memberBuilder, "id ASC")
		if err != nil {
			return errors.Wrapf(err, "find group members failed")
		}

		// 批量更新成员状态
		for _, member := range members {
			member.Status = 0 // 已退出
			_, err = l.svcCtx.ImGroupMemberModel.Update(ctx, session, member)
			if err != nil {
				l.Logger.Errorf("update member status failed: %v, userId=%d, groupId=%d", err, member.UserId, in.GroupId)
				// 继续处理其他成员
				continue
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &group.DissolveGroupResp{
		Success: true,
	}, nil
}
