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

type RecallMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 撤回消息
func NewRecallMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecallMessageLogic {
	return &RecallMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RecallMessageLogic) RecallMessage(req *types.RecallMessageReq) (resp *types.RecallMessageResp, err error) {
	// 从JWT中获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 调用RPC服务撤回消息
	rpcResp, err := l.svcCtx.MessageRpc.RecallMessage(l.ctx, &message.RecallMessageReq{
		UserId:    userId,
		MessageId: req.MessageId,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "recall message rpc failed")
	}

	return &types.RecallMessageResp{
		Success: rpcResp.Success,
	}, nil
}
