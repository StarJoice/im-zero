package logic

import (
	"context"

	"im-zero/app/message/cmd/rpc/internal/svc"
	"im-zero/app/message/cmd/rpc/message"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type MarkAsReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkAsReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkAsReadLogic {
	return &MarkAsReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 标记消息已读
func (l *MarkAsReadLogic) MarkAsRead(in *message.MarkAsReadReq) (*message.MarkAsReadResp, error) {
	// 参数验证
	if in.UserId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", in.UserId)
	}

	if len(in.ConversationId) == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "empty conversation id"), "conversation id is required")
	}

	// 使用事务确保数据一致性
	err := l.svcCtx.ImMessageModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 查找需要标记为已读的消息
		var messageIds []int64
		
		if in.MessageId > 0 {
			// 如果指定了消息ID，则标记到此消息为止的所有未读消息
			builder := l.svcCtx.ImMessageModel.SelectBuilder().
				Where("conversation_id = ? AND to_user_id = ? AND id <= ? AND status < ?", 
					in.ConversationId, in.UserId, in.MessageId, 3). // status < 3 表示未读
				OrderBy("id ASC")
			
			messages, err := l.svcCtx.ImMessageModel.FindAll(ctx, builder, "")
			if err != nil {
				return errors.Wrapf(err, "find messages to mark as read failed")
			}
			
			for _, msg := range messages {
				messageIds = append(messageIds, msg.Id)
			}
		} else {
			// 如果没有指定消息ID，则标记该会话的所有未读消息
			builder := l.svcCtx.ImMessageModel.SelectBuilder().
				Where("conversation_id = ? AND to_user_id = ? AND status < ?", 
					in.ConversationId, in.UserId, 3). // status < 3 表示未读
				OrderBy("id ASC")
			
			messages, err := l.svcCtx.ImMessageModel.FindAll(ctx, builder, "")
			if err != nil {
				return errors.Wrapf(err, "find messages to mark as read failed")
			}
			
			for _, msg := range messages {
				messageIds = append(messageIds, msg.Id)
			}
		}

		// 2. 批量更新消息状态为已读
		if len(messageIds) > 0 {
			err := l.svcCtx.ImMessageModel.UpdateMessagesStatus(ctx, session, messageIds, 3) // 3-已读
			if err != nil {
				return errors.Wrapf(err, "update messages status failed")
			}
		}

		// 3. 清零会话的未读计数
		err := l.svcCtx.ImConversationModel.ClearUnreadCount(ctx, session, in.ConversationId, in.UserId)
		if err != nil {
			return errors.Wrapf(err, "clear unread count failed")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	l.Logger.Infof("Messages marked as read: userId=%d, conversationId=%s, messageId=%d", 
		in.UserId, in.ConversationId, in.MessageId)

	return &message.MarkAsReadResp{
		Success: true,
	}, nil
}
