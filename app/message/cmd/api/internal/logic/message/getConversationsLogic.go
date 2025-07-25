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

type GetConversationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取最近对话列表
func NewGetConversationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationsLogic {
	return &GetConversationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConversationsLogic) GetConversations(req *types.GetConversationsReq) (resp *types.GetConversationsResp, err error) {
	// 从JWT中获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 调用RPC服务获取会话列表
	rpcResp, err := l.svcCtx.MessageRpc.GetConversations(l.ctx, &message.GetConversationsReq{
		UserId: userId,
		Page:   req.Page,
		Limit:  req.Limit,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get conversations rpc failed")
	}

	// 转换返回结果
	var conversations []types.Conversation
	for _, conv := range rpcResp.Conversations {
		// 转换最后一条消息
		var lastMessage types.Message
		if conv.LastMessage != nil {
			lastMessage = types.Message{
				Id:             conv.LastMessage.Id,
				FromUserId:     conv.LastMessage.FromUserId,
				ToUserId:       conv.LastMessage.ToUserId,
				ConversationId: conv.LastMessage.ConversationId,
				MessageType:    conv.LastMessage.MessageType,
				Content:        conv.LastMessage.Content,
				Extra:          conv.LastMessage.Extra,
				Status:         conv.LastMessage.Status,
				CreateTime:     conv.LastMessage.CreateTime,
				UpdateTime:     conv.LastMessage.UpdateTime,
			}
		}

		conversations = append(conversations, types.Conversation{
			Id:          conv.Id,
			Type:        conv.Type,
			Name:        conv.Name,
			Avatar:      conv.Avatar,
			LastMessage: lastMessage,
			UnreadCount: conv.UnreadCount,
			UpdateTime:  conv.UpdateTime,
		})
	}

	return &types.GetConversationsResp{
		Conversations: conversations,
		Total:         rpcResp.Total,
	}, nil
}
