package message

import (
	"context"

	"im-zero/app/message/cmd/api/internal/svc"
	"im-zero/app/message/cmd/api/internal/types"
	"im-zero/app/message/cmd/rpc/message"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取聊天记录
func NewGetChatHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatHistoryLogic {
	return &GetChatHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetChatHistoryLogic) GetChatHistory(req *types.GetChatHistoryReq) (resp *types.GetChatHistoryResp, err error) {
	// 调用RPC服务获取聊天记录
	rpcResp, err := l.svcCtx.MessageRpc.GetChatHistory(l.ctx, &message.GetChatHistoryReq{
		ConversationId: req.ConversationId,
		LastMessageId:  req.LastMessageId,
		Limit:          req.Limit,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get chat history rpc failed")
	}

	// 转换返回结果
	var messages []types.Message
	for _, msg := range rpcResp.Messages {
		messages = append(messages, types.Message{
			Id:             msg.Id,
			FromUserId:     msg.FromUserId,
			ToUserId:       msg.ToUserId,
			ConversationId: msg.ConversationId,
			MessageType:    msg.MessageType,
			Content:        msg.Content,
			Extra:          msg.Extra,
			Status:         msg.Status,
			CreateTime:     msg.CreateTime,
			UpdateTime:     msg.UpdateTime,
		})
	}

	return &types.GetChatHistoryResp{
		Messages: messages,
		HasMore:  rpcResp.HasMore,
	}, nil
}
