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

type SendMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 发送消息
func NewSendMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMessageLogic {
	return &SendMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendMessageLogic) SendMessage(req *types.SendMessageReq) (resp *types.SendMessageResp, err error) {
	// 从JWT中获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 调用RPC服务发送消息
	rpcResp, err := l.svcCtx.MessageRpc.SendMessage(l.ctx, &message.SendMessageReq{
		FromUserId:  userId,
		ToUserId:    req.ToUserId,
		MessageType: req.MessageType,
		Content:     req.Content,
		Extra:       req.Extra,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "send message rpc failed")
	}

	// 转换返回结果
	return &types.SendMessageResp{
		Message: types.Message{
			Id:             rpcResp.Message.Id,
			FromUserId:     rpcResp.Message.FromUserId,
			ToUserId:       rpcResp.Message.ToUserId,
			ConversationId: rpcResp.Message.ConversationId,
			MessageType:    rpcResp.Message.MessageType,
			Content:        rpcResp.Message.Content,
			Extra:          rpcResp.Message.Extra,
			Status:         rpcResp.Message.Status,
			CreateTime:     rpcResp.Message.CreateTime,
			UpdateTime:     rpcResp.Message.UpdateTime,
		},
	}, nil
}
