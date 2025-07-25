package logic

import (
	"context"

	"im-zero/app/friend/cmd/rpc/friend"
	"im-zero/app/friend/cmd/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckBlockedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckBlockedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckBlockedLogic {
	return &CheckBlockedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 检查是否被拉黑
func (l *CheckBlockedLogic) CheckBlocked(in *friend.CheckBlockedReq) (*friend.CheckBlockedResp, error) {
	// todo: add your logic here and delete this line

	return &friend.CheckBlockedResp{}, nil
}
