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

type HandleFriendRequestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 处理好友请求
func NewHandleFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleFriendRequestLogic {
	return &HandleFriendRequestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleFriendRequestLogic) HandleFriendRequest(req *types.HandleFriendRequestReq) (resp *types.HandleFriendRequestResp, err error) {
	// 参数验证
	if req.RequestId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid request id"), "requestId=%d", req.RequestId)
	}
	if req.Action != 1 && req.Action != 2 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid action"), "action must be 1(accept) or 2(reject), got=%d", req.Action)
	}
	if req.Action == 1 && len(req.Remark) > 50 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "remark too long"), "remark length=%d", len(req.Remark))
	}

	// 获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.UNAUTHORIZED, "user not login"), "userId is empty")
	}

	// 调用RPC服务处理好友请求
	rpcResp, err := l.svcCtx.FriendRpc.HandleFriendRequest(l.ctx, &friend.HandleFriendRequestReq{
		RequestId: req.RequestId,
		UserId:    userId,
		Action:    req.Action,
		Remark:    req.Remark,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "handle friend request rpc failed")
	}

	// 记录操作日志
	actionStr := "rejected"
	if req.Action == 1 {
		actionStr = "accepted"
	}
	l.Logger.Infof("Friend request %s successfully: userId=%d, requestId=%d", actionStr, userId, req.RequestId)

	return &types.HandleFriendRequestResp{
		Success: rpcResp.Success,
	}, nil
}
