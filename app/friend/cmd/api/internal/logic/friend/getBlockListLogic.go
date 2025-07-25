package friend

import (
	"context"

	"im-zero/app/friend/cmd/api/internal/svc"
	"im-zero/app/friend/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlockListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取黑名单
func NewGetBlockListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlockListLogic {
	return &GetBlockListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBlockListLogic) GetBlockList(req *types.GetBlockListReq) (resp *types.GetBlockListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
