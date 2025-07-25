package logic

import (
	"context"
	"im-zero/pkg/tool"

	"im-zero/app/group/cmd/rpc/group"
	"im-zero/app/group/cmd/rpc/internal/svc"
	"im-zero/app/group/model"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupInfoLogic {
	return &GetGroupInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取群组信息
func (l *GetGroupInfoLogic) GetGroupInfo(in *group.GetGroupInfoReq) (*group.GetGroupInfoResp, error) {
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
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.NOT_FOUND, "group is not active"), "groupId=%d, status=%d", in.GroupId, groupInfo.Status)
	}

	// 检查用户是否在群中
	member, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, in.UserId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "user not in group"), "userId=%d, groupId=%d", in.UserId, in.GroupId)
		}
		return nil, errors.Wrapf(err, "check group member failed")
	}

	// 检查成员状态
	if member.Status != 1 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "user not active in group"), "userId=%d, status=%d", in.UserId, member.Status)
	}

	// 返回群组信息
	return &group.GetGroupInfoResp{
		Group: &group.GroupInfo{
			Id:                groupInfo.Id,
			Name:              groupInfo.Name,
			Avatar:            groupInfo.Avatar.String,
			Description:       groupInfo.Description.String,
			Notice:            groupInfo.Notice.String,
			OwnerId:           groupInfo.OwnerId,
			MemberCount:       int32(groupInfo.MemberCount),
			MaxMembers:        int32(groupInfo.MaxMembers),
			Status:            int32(groupInfo.Status),
			IsPrivate:         tool.Int64ToBool(groupInfo.IsPrivate),
			JoinApproval:      tool.Int64ToBool(groupInfo.JoinApproval),
			AllowInvite:       tool.Int64ToBool(groupInfo.AllowInvite),
			AllowMemberModify: tool.Int64ToBool(groupInfo.AllowMemberModify),
			CreateTime:        groupInfo.CreateTime.Unix(),
			UpdateTime:        groupInfo.UpdateTime.Unix(),
		},
		MyRole: int32(member.Role),
	}, nil
}
