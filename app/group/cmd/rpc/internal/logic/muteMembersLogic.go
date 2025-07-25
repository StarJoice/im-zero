package logic

import (
	"context"
	"database/sql"
	"time"

	"im-zero/app/group/cmd/rpc/group"
	"im-zero/app/group/cmd/rpc/internal/svc"
	"im-zero/app/group/model"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type MuteMembersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMuteMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MuteMembersLogic {
	return &MuteMembersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 禁言成员
func (l *MuteMembersLogic) MuteMembers(in *group.MuteMembersReq) (*group.MuteMembersResp, error) {
	// 参数验证
	if in.GroupId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid group id"), "groupId=%d", in.GroupId)
	}
	if in.OperatorId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid operator id"), "operatorId=%d", in.OperatorId)
	}
	if len(in.UserIds) == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "user ids is required"), "userIds is empty")
	}
	if in.Duration < 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid duration"), "duration=%d", in.Duration)
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

	// 只有群主和管理员可以禁言
	if operatorMember.Role != 3 && operatorMember.Role != 2 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "only owner and admin can mute members"), "operatorId=%d, role=%d", in.OperatorId, operatorMember.Role)
	}

	var successCount int32
	var failedUsers []int64

	// 使用事务处理批量禁言
	err = l.svcCtx.ImGroupModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, userId := range in.UserIds {
			// 不能禁言自己
			if userId == in.OperatorId {
				failedUsers = append(failedUsers, userId)
				l.Logger.Infof("cannot mute self, operatorId=%d", in.OperatorId)
				continue
			}

			// 检查目标用户是否在群中
			targetMember, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(ctx, in.GroupId, userId)
			if err != nil {
				if errors.Is(err, model.ErrNotFound) {
					// 用户不在群中
					failedUsers = append(failedUsers, userId)
					l.Logger.Infof("user %d not in group %d", userId, in.GroupId)
					continue
				}
				return errors.Wrapf(err, "check target member failed")
			}

			// 检查目标用户状态
			if targetMember.Status != 1 {
				failedUsers = append(failedUsers, userId)
				l.Logger.Infof("user %d not active in group %d", userId, in.GroupId)
				continue
			}

			// 检查权限：管理员不能禁言群主和其他管理员
			if operatorMember.Role == 2 { // 操作者是管理员
				if targetMember.Role == 3 { // 不能禁言群主
					failedUsers = append(failedUsers, userId)
					l.Logger.Infof("admin cannot mute owner, operatorId=%d, targetUserId=%d", in.OperatorId, userId)
					continue
				}
				if targetMember.Role == 2 { // 不能禁言其他管理员
					failedUsers = append(failedUsers, userId)
					l.Logger.Infof("admin cannot mute other admin, operatorId=%d, targetUserId=%d", in.OperatorId, userId)
					continue
				}
			}

			// 设置禁言时间
			if in.Duration == 0 {
				// 解除禁言
				targetMember.MuteEndTime = sql.NullTime{Valid: false}
			} else {
				// 设置禁言结束时间
				muteEndTime := time.Now().Add(time.Duration(in.Duration) * time.Second)
				targetMember.MuteEndTime = sql.NullTime{
					Time:  muteEndTime,
					Valid: true,
				}
			}

			// 更新成员禁言状态
			_, err = l.svcCtx.ImGroupMemberModel.Update(ctx, session, targetMember)
			if err != nil {
				l.Logger.Errorf("update member mute status failed: %v, userId=%d, groupId=%d", err, userId, in.GroupId)
				failedUsers = append(failedUsers, userId)
				continue
			}

			successCount++
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &group.MuteMembersResp{
		SuccessCount: successCount,
		FailedUsers:  failedUsers,
	}, nil
}
