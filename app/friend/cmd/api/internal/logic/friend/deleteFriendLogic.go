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

type DeleteFriendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除好友
func NewDeleteFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFriendLogic {
	return &DeleteFriendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteFriendLogic) DeleteFriend(req *types.DeleteFriendReq) (resp *types.DeleteFriendResp, err error) {
	// 参数验证
	if req.FriendUserId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid friend user id"), "friendUserId=%d", req.FriendUserId)
	}

	// 获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.UNAUTHORIZED, "user not login"), "userId is empty")
	}

	// 不能删除自己
	if userId == req.FriendUserId {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "cannot delete yourself"), "userId=%d", userId)
	}

	// 调用RPC服务删除好友
	rpcResp, err := l.svcCtx.FriendRpc.DeleteFriend(l.ctx, &friend.DeleteFriendReq{
		UserId:       userId,
		FriendUserId: req.FriendUserId,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "delete friend rpc failed")
	}

	// 记录操作日志
	l.Logger.Infof("Delete friend successfully: userId=%d, friendUserId=%d", userId, req.FriendUserId)

	return &types.DeleteFriendResp{
		Success: rpcResp.Success,
	}, nil
}
