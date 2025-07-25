package message

import (
	"context"

	"im-zero/app/message/cmd/api/internal/svc"
	"im-zero/app/message/cmd/api/internal/types"
	"im-zero/app/message/cmd/rpc/message"
	"im-zero/pkg/ctxdata"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除消息
func NewDeleteMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMessageLogic {
	return &DeleteMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteMessageLogic) DeleteMessage(req *types.DeleteMessageReq) (resp *types.DeleteMessageResp, err error) {
	// 从JWT中获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 调用RPC服务删除消息
	rpcResp, err := l.svcCtx.MessageRpc.DeleteMessage(l.ctx, &message.DeleteMessageReq{
		UserId:    userId,
		MessageId: req.MessageId,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "delete message rpc failed")
	}

	return &types.DeleteMessageResp{
		Success: rpcResp.Success,
	}, nil
}
