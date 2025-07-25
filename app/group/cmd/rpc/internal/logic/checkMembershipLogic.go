package logic

import (
	"context"

	"im-zero/app/group/cmd/rpc/group"
	"im-zero/app/group/cmd/rpc/internal/svc"
	"im-zero/app/group/model"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type CheckMembershipLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckMembershipLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckMembershipLogic {
	return &CheckMembershipLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 检查用户是否在群中
func (l *CheckMembershipLogic) CheckMembership(in *group.CheckMembershipReq) (*group.CheckMembershipResp, error) {
	// 参数验证
	if in.GroupId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid group id"), "groupId=%d", in.GroupId)
	}
	if in.UserId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", in.UserId)
	}

	// 查找群成员
	member, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, in.UserId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			// 不是群成员
			return &group.CheckMembershipResp{
				IsMember: false,
				Role:     0,
			}, nil
		}
		return nil, errors.Wrapf(err, "check group member failed")
	}

	// 检查成员状态
	if member.Status != 1 {
		// 成员状态不正常（已退出或被踢出）
		return &group.CheckMembershipResp{
			IsMember: false,
			Role:     0,
		}, nil
	}

	// 是正常的群成员
	return &group.CheckMembershipResp{
		IsMember: true,
		Role:     int32(member.Role),
	}, nil
}
