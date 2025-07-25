package friend

import (
	"context"

	"im-zero/app/friend/cmd/api/internal/svc"
	"im-zero/app/friend/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnblockUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 取消拉黑
func NewUnblockUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnblockUserLogic {
	return &UnblockUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnblockUserLogic) UnblockUser(req *types.UnblockUserReq) (resp *types.UnblockUserResp, err error) {
	// todo: add your logic here and delete this line

	return
}
