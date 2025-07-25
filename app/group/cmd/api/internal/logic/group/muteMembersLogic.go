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

type MuteMembersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 禁言群成员
func NewMuteMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MuteMembersLogic {
	return &MuteMembersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MuteMembersLogic) MuteMembers(req *types.MuteMembersReq) (resp *types.MuteMembersResp, err error) {
	// 参数验证
	if req.GroupId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid group id"), "groupId=%d", req.GroupId)
	}
	if len(req.UserIds) == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "user ids is required"), "userIds is empty")
	}
	if req.Duration < 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid duration"), "duration=%d", req.Duration)
	}

	// 获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.UNAUTHORIZED, "user not login"), "userId is empty")
	}

	// 调用RPC服务禁言成员
	rpcResp, err := l.svcCtx.GroupRpc.MuteMembers(l.ctx, &group.MuteMembersReq{
		GroupId:    req.GroupId,
		OperatorId: userId,
		UserIds:    req.UserIds,
		Duration:   req.Duration,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "mute members from rpc failed")
	}

	return &types.MuteMembersResp{
		SuccessCount: rpcResp.SuccessCount,
		FailedUsers:  rpcResp.FailedUsers,
	}, nil
}
