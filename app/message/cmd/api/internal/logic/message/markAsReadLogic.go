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

type MarkAsReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 标记消息为已读
func NewMarkAsReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkAsReadLogic {
	return &MarkAsReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkAsReadLogic) MarkAsRead(req *types.MarkAsReadReq) (resp *types.MarkAsReadResp, err error) {
	// 从JWT中获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 调用RPC服务标记已读
	rpcResp, err := l.svcCtx.MessageRpc.MarkAsRead(l.ctx, &message.MarkAsReadReq{
		UserId:         userId,
		ConversationId: req.ConversationId,
		MessageId:      req.MessageId,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "mark as read rpc failed")
	}

	return &types.MarkAsReadResp{
		Success: rpcResp.Success,
	}, nil
}
