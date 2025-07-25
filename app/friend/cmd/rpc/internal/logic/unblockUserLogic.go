package logic

import (
	"context"

	"im-zero/app/friend/cmd/rpc/friend"
	"im-zero/app/friend/cmd/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnblockUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnblockUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnblockUserLogic {
	return &UnblockUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 取消拉黑
func (l *UnblockUserLogic) UnblockUser(in *friend.UnblockUserReq) (*friend.UnblockUserResp, error) {
	// todo: add your logic here and delete this line

	return &friend.UnblockUserResp{}, nil
}
