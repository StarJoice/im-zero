package logic

import (
	"context"
	"database/sql"
	"im-zero/pkg/tool"
	"time"

	"im-zero/app/group/cmd/rpc/group"
	"im-zero/app/group/cmd/rpc/internal/svc"
	"im-zero/app/group/model"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGroupLogic {
	return &CreateGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建群组
func (l *CreateGroupLogic) CreateGroup(in *group.CreateGroupReq) (*group.CreateGroupResp, error) {
	// 参数验证
	if in.OwnerId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid owner id"), "ownerId=%d", in.OwnerId)
	}
	if len(in.Name) == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "group name is required"), "name is empty")
	}
	if len(in.Name) > 100 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "group name too long"), "name length=%d", len(in.Name))
	}
	if len(in.Description) > 500 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "group description too long"), "description length=%d", len(in.Description))
	}
	
	// 检查初始成员数量限制
	totalMembers := len(in.MemberIds) + 1 // +1 for owner
	if totalMembers > 500 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.GROUP_MEMBER_FULL, "too many initial members"), "memberCount=%d", totalMembers)
	}

	// 使用事务创建群组
	var groupInfo *model.ImGroup
	var memberCount int32

	err := l.svcCtx.ImGroupModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 创建群组
		newGroup := &model.ImGroup{
			Name:              in.Name,
			Avatar:            sql.NullString{String: in.Avatar, Valid: in.Avatar != ""},
			Description:       sql.NullString{String: in.Description, Valid: in.Description != ""},
			OwnerId:           in.OwnerId,
			IsPrivate:         tool.BoolToInt64(in.IsPrivate),
			JoinApproval:      tool.BoolToInt64(in.JoinApproval),
			AllowInvite:       1,   // 默认允许邀请
			AllowMemberModify: 0,   // 默认不允许成员修改
			Status:            1,   // 正常状态
			MaxMembers:        500, // 默认最大成员数
		}

		_, err := l.svcCtx.ImGroupModel.Insert(ctx, session, newGroup)
		if err != nil {
			return errors.Wrapf(err, "insert group failed")
		}

		// 添加群主为成员
		ownerMember := &model.ImGroupMember{
			GroupId:    newGroup.Id,
			UserId:     in.OwnerId,
			Role:       3, // 群主
			Status:     1, // 正常状态
			JoinSource: 1, // 创建方式
			JoinTime:   time.Now(),
		}

		_, err = l.svcCtx.ImGroupMemberModel.Insert(ctx, session, ownerMember)
		if err != nil {
			return errors.Wrapf(err, "insert owner member failed")
		}

		memberCount = 1

		// 添加其他成员
		for _, memberId := range in.MemberIds {
			if memberId == in.OwnerId {
				continue // 跳过群主
			}

			member := &model.ImGroupMember{
				GroupId:    newGroup.Id,
				UserId:     memberId,
				Role:       1, // 普通成员
				Status:     1, // 正常状态
				JoinSource: 1, // 邀请加入
				InviterId:  sql.NullInt64{Int64: in.OwnerId, Valid: true},
				JoinTime:   time.Now(),
			}

			_, err = l.svcCtx.ImGroupMemberModel.Insert(ctx, session, member)
			if err != nil {
				l.Logger.Errorf("insert member %d failed: %v", memberId, err)
				continue // 继续添加其他成员
			}
			memberCount++
		}

		// 更新群组成员数量
		newGroup.MemberCount = int64(memberCount)
		_, err = l.svcCtx.ImGroupModel.Update(ctx, session, newGroup)
		if err != nil {
			return errors.Wrapf(err, "update group member count failed")
		}

		groupInfo = newGroup
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 返回群组信息
	return &group.CreateGroupResp{
		Group: &group.GroupInfo{
			Id:                groupInfo.Id,
			Name:              groupInfo.Name,
			Avatar:            groupInfo.Avatar.String,
			Description:       groupInfo.Description.String,
			OwnerId:           groupInfo.OwnerId,
			MemberCount:       memberCount,
			MaxMembers:        int32(groupInfo.MaxMembers),
			Status:            int32(groupInfo.Status),
			IsPrivate:         tool.Int64ToBool(groupInfo.IsPrivate),
			JoinApproval:      tool.Int64ToBool(groupInfo.JoinApproval),
			AllowInvite:       tool.Int64ToBool(groupInfo.AllowInvite),
			AllowMemberModify: tool.Int64ToBool(groupInfo.AllowMemberModify),
			CreateTime:        groupInfo.CreateTime.Unix(),
			UpdateTime:        groupInfo.UpdateTime.Unix(),
		},
	}, nil
}
