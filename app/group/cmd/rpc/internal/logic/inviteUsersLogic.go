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

type InviteUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInviteUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InviteUsersLogic {
	return &InviteUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 邀请用户入群
func (l *InviteUsersLogic) InviteUsers(in *group.InviteUsersReq) (*group.InviteUsersResp, error) {
	// 参数验证
	if in.GroupId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid group id"), "groupId=%d", in.GroupId)
	}
	if in.InviterId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid inviter id"), "inviterId=%d", in.InviterId)
	}
	if len(in.UserIds) == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "user ids is required"), "userIds is empty")
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

	// 检查邀请者权限
	inviterMember, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, in.InviterId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "inviter not in group"), "inviterId=%d, groupId=%d", in.InviterId, in.GroupId)
		}
		return nil, errors.Wrapf(err, "check inviter member failed")
	}

	// 检查邀请者状态
	if inviterMember.Status != 1 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "inviter not active in group"), "inviterId=%d, status=%d", in.InviterId, inviterMember.Status)
	}

	// 检查是否允许邀请（群主和管理员总是可以邀请，普通成员需要检查群设置）
	if inviterMember.Role == 1 && groupInfo.AllowInvite != 1 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "group not allow member invite"), "groupId=%d", in.GroupId)
	}

	var successCount int32
	var failedUsers []int64

	// 使用事务处理批量邀请
	err = l.svcCtx.ImGroupModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, userId := range in.UserIds {
			// 验证用户ID
			if userId <= 0 {
				failedUsers = append(failedUsers, userId)
				l.Logger.Infof("invalid user id: %d", userId)
				continue
			}

			// 检查用户是否已经在群中
			existingMember, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(ctx, in.GroupId, userId)
			if err == nil {
				if existingMember.Status == 1 {
					// 用户已经在群中且状态正常
					failedUsers = append(failedUsers, userId)
					l.Logger.Infof("user %d already in group %d with active status", userId, in.GroupId)
					continue
				} else if existingMember.Status == 2 {
					// 用户曾被踢出，需要特殊处理
					failedUsers = append(failedUsers, userId)
					l.Logger.Infof("user %d was removed from group %d, cannot reinvite directly", userId, in.GroupId)
					continue
				}
				// 如果状态是0（已退出），可以重新邀请，继续处理
			} else if !errors.Is(err, model.ErrNotFound) {
				// 数据库查询出错
				l.Logger.Errorf("check existing member failed: %v, userId=%d", err, userId)
				failedUsers = append(failedUsers, userId)
				continue
			}

			// 检查群成员数量限制
			if groupInfo.MemberCount >= groupInfo.MaxMembers {
				// 群已满
				failedUsers = append(failedUsers, userId)
				l.Logger.Infof("group %d is full (current: %d, max: %d), cannot invite user %d", 
					in.GroupId, groupInfo.MemberCount, groupInfo.MaxMembers, userId)
				continue
			}

			// 如果是重新邀请已退出的用户，更新现有记录
			if existingMember != nil && existingMember.Status == 0 {
				existingMember.Status = 1
				existingMember.Role = 1
				existingMember.JoinSource = 1
				existingMember.InviterId = sql.NullInt64{Int64: in.InviterId, Valid: true}
				existingMember.JoinTime = time.Now()

				_, err = l.svcCtx.ImGroupMemberModel.Update(ctx, session, existingMember)
				if err != nil {
					l.Logger.Errorf("update existing member failed: %v, userId=%d, groupId=%d", err, userId, in.GroupId)
					failedUsers = append(failedUsers, userId)
					continue
				}
			} else {
				// 创建新成员记录
				newMember := &model.ImGroupMember{
					GroupId:    in.GroupId,
					UserId:     userId,
					Role:       1, // 普通成员
					Status:     1, // 正常状态
					JoinSource: 1, // 邀请加入
					InviterId:  sql.NullInt64{Int64: in.InviterId, Valid: true},
					JoinTime:   time.Now(),
				}

				_, err = l.svcCtx.ImGroupMemberModel.Insert(ctx, session, newMember)
				if err != nil {
					l.Logger.Errorf("insert group member failed: %v, userId=%d, groupId=%d", err, userId, in.GroupId)
					failedUsers = append(failedUsers, userId)
					continue
				}
			}

			successCount++
			// 实时更新群组成员数，避免并发问题
			groupInfo.MemberCount++
		}

		// 最终更新群组成员数量
		if successCount > 0 {
			_, err = l.svcCtx.ImGroupModel.Update(ctx, session, groupInfo)
			if err != nil {
				return errors.Wrapf(err, "update group member count failed")
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &group.InviteUsersResp{
		SuccessCount: successCount,
		FailedUsers:  failedUsers,
	}, nil
}
