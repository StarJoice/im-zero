package logic

import (
	"context"

	"im-zero/app/message/cmd/rpc/internal/svc"
	"im-zero/app/message/cmd/rpc/message"
	"im-zero/pkg/wsmanager"

	"github.com/zeromicro/go-zero/core/logx"
)

type PushMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPushMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PushMessageLogic {
	return &PushMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 推送消息给用户
func (l *PushMessageLogic) PushMessage(in *message.PushMessageReq) (*message.PushMessageResp, error) {
	// 参数验证
	if in.UserId <= 0 {
		l.Logger.Errorf("Invalid user id: %d", in.UserId)
		return &message.PushMessageResp{Success: false}, nil
	}

	if in.Message == nil {
		l.Logger.Errorf("Message is nil for user: %d", in.UserId)
		return &message.PushMessageResp{Success: false}, nil
	}

	// 获取WebSocket连接管理器
	hub := wsmanager.GetHub()
	
	// 检查用户是否在线
	isOnline := hub.IsUserOnline(in.UserId)
	
	if isOnline {
		// 用户在线，通过WebSocket推送消息
		success := hub.SendToUser(in.UserId, "new_message", map[string]interface{}{
			"id":             in.Message.Id,
			"from_user_id":   in.Message.FromUserId,
			"to_user_id":     in.Message.ToUserId,
			"conversation_id": in.Message.ConversationId,
			"message_type":   in.Message.MessageType,
			"content":        in.Message.Content,
			"extra":          in.Message.Extra,
			"status":         in.Message.Status,
			"create_time":    in.Message.CreateTime,
			"update_time":    in.Message.UpdateTime,
		})
		
		if success {
			// 推送成功，更新消息状态为已送达
			err := l.svcCtx.ImMessageModel.UpdateMessageStatus(l.ctx, in.Message.Id, 2) // 2-已送达
			if err != nil {
				l.Logger.Errorf("Update message status to delivered failed: msgId=%d, err=%v", in.Message.Id, err)
			}
			
			l.Logger.Infof("Message pushed successfully to online user %d, msgId=%d", in.UserId, in.Message.Id)
			return &message.PushMessageResp{Success: true}, nil
		} else {
			l.Logger.Errorf("Failed to push message to online user %d via WebSocket", in.UserId)
		}
	} else {
		l.Logger.Infof("User %d is offline, message will be delivered when user comes online", in.UserId)
		// 用户离线，消息保持在数据库中，状态为已发送(1)
		// 当用户上线时，可以通过获取聊天记录接口获取未读消息
	}

	// 无论是否在线都返回成功，因为消息已经保存到数据库
	// 离线消息会在用户上线后通过获取聊天记录等方式获取
	return &message.PushMessageResp{Success: true}, nil
}
