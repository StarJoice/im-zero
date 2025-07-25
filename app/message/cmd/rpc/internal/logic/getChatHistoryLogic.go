package logic

import (
	"context"

	"im-zero/app/message/cmd/rpc/internal/svc"
	"im-zero/app/message/cmd/rpc/message"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChatHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatHistoryLogic {
	return &GetChatHistoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取聊天记录
func (l *GetChatHistoryLogic) GetChatHistory(in *message.GetChatHistoryReq) (*message.GetChatHistoryResp, error) {
	// 参数验证
	if len(in.ConversationId) == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "empty conversation id"), "conversation id is required")
	}

	limit := int(in.Limit)
	if limit <= 0 || limit > 100 {
		limit = 20 // 默认20条
	}

	// 查询消息记录（排除已删除的消息）
	builder := l.svcCtx.ImMessageModel.SelectBuilder().
		Where("conversation_id = ? AND status != ?", in.ConversationId, 5) // 排除已删除(5)的消息
	
	if in.LastMessageId > 0 {
		builder = builder.Where("id < ?", in.LastMessageId)
	}
	
	messages, err := l.svcCtx.ImMessageModel.FindAll(l.ctx, builder.OrderBy("id DESC").Limit(uint64(limit+1)), "")
	if err != nil {
		return nil, errors.Wrapf(err, "find messages failed")
	}

	// 判断是否还有更多数据
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit] // 去掉多查的一条
	}

	// 转换数据格式，并按时间正序排列（最老的在前面）
	var messageInfos []*message.MessageInfo
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		
		// 根据消息状态决定显示内容
		content := msg.Content
		if msg.Status == 4 { // 撤回状态
			content = "[消息已撤回]"
		}
		
		messageInfos = append(messageInfos, &message.MessageInfo{
			Id:             msg.Id,
			FromUserId:     msg.FromUserId,
			ToUserId:       msg.ToUserId,
			ConversationId: msg.ConversationId,
			MessageType:    int32(msg.MessageType),
			Content:        content,
			Extra:          msg.Extra.String,
			Status:         int32(msg.Status),
			CreateTime:     msg.CreateTime.Unix(),
			UpdateTime:     msg.UpdateTime.Unix(),
		})
	}

	return &message.GetChatHistoryResp{
		Messages: messageInfos,
		HasMore:  hasMore,
	}, nil
}
