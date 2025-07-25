package logic

import (
	"context"

	"im-zero/app/friend/cmd/rpc/friend"
	"im-zero/app/friend/cmd/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlockListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBlockListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlockListLogic {
	return &GetBlockListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取黑名单
func (l *GetBlockListLogic) GetBlockList(in *friend.GetBlockListReq) (*friend.GetBlockListResp, error) {
	// todo: add your logic here and delete this line

	return &friend.GetBlockListResp{}, nil
}
