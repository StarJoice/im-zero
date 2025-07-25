package friend

import (
	"context"

	"im-zero/app/friend/cmd/api/internal/svc"
	"im-zero/app/friend/cmd/api/internal/types"

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
	// todo: add your logic here and delete this line

	return
}
