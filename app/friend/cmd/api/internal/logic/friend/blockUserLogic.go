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

type BlockUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 拉黑用户
func NewBlockUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlockUserLogic {
	return &BlockUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BlockUserLogic) BlockUser(req *types.BlockUserReq) (resp *types.BlockUserResp, err error) {
	// 参数验证
	if req.UserId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", req.UserId)
	}
	if len(req.Reason) > 100 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "reason too long"), "reason length=%d", len(req.Reason))
	}

	// 获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.UNAUTHORIZED, "user not login"), "userId is empty")
	}

	// 不能拉黑自己
	if userId == req.UserId {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "cannot block yourself"), "userId=%d", userId)
	}

	// 调用RPC服务拉黑用户
	rpcResp, err := l.svcCtx.FriendRpc.BlockUser(l.ctx, &friend.BlockUserReq{
		UserId:        userId,
		BlockedUserId: req.UserId,
		Reason:        req.Reason,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "block user rpc failed")
	}

	// 记录操作日志
	l.Logger.Infof("Block user successfully: userId=%d, blockedUserId=%d, reason=%s", userId, req.UserId, req.Reason)

	return &types.BlockUserResp{
		Success: rpcResp.Success,
	}, nil
}
