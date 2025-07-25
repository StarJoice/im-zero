package logic

import (
	"context"

	"im-zero/app/friend/cmd/rpc/friend"
	"im-zero/app/friend/cmd/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckFriendshipLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckFriendshipLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckFriendshipLogic {
	return &CheckFriendshipLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 检查好友关系
func (l *CheckFriendshipLogic) CheckFriendship(in *friend.CheckFriendshipReq) (*friend.CheckFriendshipResp, error) {
	// todo: add your logic here and delete this line

	return &friend.CheckFriendshipResp{}, nil
}
