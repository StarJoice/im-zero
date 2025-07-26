package friend

import (
	"context"

	"im-zero/app/friend/cmd/api/internal/svc"
	"im-zero/app/friend/cmd/api/internal/types"
	"im-zero/app/friend/cmd/rpc/friend"
	"im-zero/pkg/ctxdata"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type UnblockUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 取消拉黑
func NewUnblockUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnblockUserLogic {
	return &UnblockUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnblockUserLogic) UnblockUser(req *types.UnblockUserReq) (resp *types.UnblockUserResp, err error) {
	// 参数验证
	if req.UserId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", req.UserId)
	}

	// 获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.UNAUTHORIZED, "user not login"), "userId is empty")
	}

	// 不能取消拉黑自己
	if userId == req.UserId {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "cannot unblock yourself"), "userId=%d", userId)
	}

	// 调用RPC服务取消拉黑
	rpcResp, err := l.svcCtx.FriendRpc.UnblockUser(l.ctx, &friend.UnblockUserReq{
		UserId:        userId,
		BlockedUserId: req.UserId,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unblock user rpc failed")
	}

	// 记录操作日志
	l.Logger.Infof("Unblock user successfully: userId=%d, unblockedUserId=%d", userId, req.UserId)

	return &types.UnblockUserResp{
		Success: rpcResp.Success,
	}, nil
}
