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

type RemoveMembersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 踢出群成员
func NewRemoveMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveMembersLogic {
	return &RemoveMembersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveMembersLogic) RemoveMembers(req *types.RemoveMembersReq) (resp *types.RemoveMembersResp, err error) {
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

	// 调用RPC服务移除成员
	rpcResp, err := l.svcCtx.GroupRpc.RemoveMembers(l.ctx, &group.RemoveMembersReq{
		GroupId:    req.GroupId,
		OperatorId: userId,
		UserIds:    req.UserIds,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "remove members from rpc failed")
	}

	return &types.RemoveMembersResp{
		SuccessCount: rpcResp.SuccessCount,
		FailedUsers:  rpcResp.FailedUsers,
	}, nil
}
