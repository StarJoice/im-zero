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

type DissolveGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 解散群组
func NewDissolveGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DissolveGroupLogic {
	return &DissolveGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DissolveGroupLogic) DissolveGroup(req *types.DissolveGroupReq) (resp *types.DissolveGroupResp, err error) {
	// 参数验证
	if req.GroupId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid group id"), "groupId=%d", req.GroupId)
	}

	// 获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.UNAUTHORIZED, "user not login"), "userId is empty")
	}

	// 调用RPC服务解散群组
	rpcResp, err := l.svcCtx.GroupRpc.DissolveGroup(l.ctx, &group.DissolveGroupReq{
		GroupId:    req.GroupId,
		OperatorId: userId,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "dissolve group from rpc failed")
	}

	return &types.DissolveGroupResp{
		Success: rpcResp.Success,
	}, nil
}
