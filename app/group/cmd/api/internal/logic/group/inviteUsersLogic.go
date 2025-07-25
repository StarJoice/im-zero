package group

import (
	"context"

	"im-zero/app/group/cmd/api/internal/svc"
	"im-zero/app/group/cmd/api/internal/types"
	"im-zero/app/group/cmd/rpc/group"
	"im-zero/pkg/ctxdata"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type InviteUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 邀请用户入群
func NewInviteUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InviteUsersLogic {
	return &InviteUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InviteUsersLogic) InviteUsers(req *types.InviteUsersReq) (resp *types.InviteUsersResp, err error) {
	// 参数验证
	if req.GroupId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid group id"), "groupId=%d", req.GroupId)
	}
	if len(req.UserIds) == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "user ids is required"), "userIds is empty")
	}

	// 获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.UNAUTHORIZED, "user not login"), "userId is empty")
	}

	// 调用RPC服务邀请用户
	rpcResp, err := l.svcCtx.GroupRpc.InviteUsers(l.ctx, &group.InviteUsersReq{
		GroupId:   req.GroupId,
		InviterId: userId,
		UserIds:   req.UserIds,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "invite users from rpc failed")
	}

	return &types.InviteUsersResp{
		SuccessCount: rpcResp.SuccessCount,
		FailedUsers:  rpcResp.FailedUsers,
	}, nil
}
