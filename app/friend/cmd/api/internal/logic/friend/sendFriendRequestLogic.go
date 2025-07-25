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

type SendFriendRequestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 发送好友请求
func NewSendFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendFriendRequestLogic {
	return &SendFriendRequestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendFriendRequestLogic) SendFriendRequest(req *types.SendFriendRequestReq) (resp *types.SendFriendRequestResp, err error) {
	// 参数验证
	if req.ToUserId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "toUserId=%d", req.ToUserId)
	}
	if len(req.Message) > 200 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "message too long"), "message length=%d", len(req.Message))
	}

	// 获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.UNAUTHORIZED, "user not login"), "userId is empty")
	}

	// 不能添加自己为好友
	if userId == req.ToUserId {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "cannot add yourself as friend"), "userId=%d", userId)
	}

	// 调用RPC服务发送好友请求
	rpcResp, err := l.svcCtx.FriendRpc.SendFriendRequest(l.ctx, &friend.SendFriendRequestReq{
		FromUserId: userId,
		ToUserId:   req.ToUserId,
		Message:    req.Message,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "send friend request rpc failed")
	}

	// 记录操作日志
	l.Logger.Infof("Friend request sent successfully: from=%d, to=%d, requestId=%d", userId, req.ToUserId, rpcResp.RequestId)

	return &types.SendFriendRequestResp{
		RequestId: rpcResp.RequestId,
	}, nil
}
