package logic

import (
	"context"
	"im-zero/app/message/model"

	"im-zero/app/message/cmd/rpc/internal/svc"
	"im-zero/app/message/cmd/rpc/message"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DeleteMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMessageLogic {
	return &DeleteMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除消息
func (l *DeleteMessageLogic) DeleteMessage(in *message.DeleteMessageReq) (*message.DeleteMessageResp, error) {
	// 参数验证
	if in.UserId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", in.UserId)
	}

	if in.MessageId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid message id"), "messageId=%d", in.MessageId)
	}

	// 使用事务确保数据一致性
	err := l.svcCtx.ImMessageModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 查找消息
		msg, err := l.svcCtx.ImMessageModel.FindOne(ctx, in.MessageId)
		if err != nil {
			return errors.Wrapf(err, "find message failed")
		}

		// 2. 验证权限（发送者和接收者都可以删除消息，但删除是单向的）
		if msg.FromUserId != in.UserId && msg.ToUserId != in.UserId {
			return errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "permission denied"),
				"user %d cannot delete message between %d and %d", in.UserId, msg.FromUserId, msg.ToUserId)
		}

		// 3. 检查消息是否已删除
		if msg.Status == 5 { // 5-删除
			return errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "message already deleted"), "messageId=%d", in.MessageId)
		}

		// 4. 软删除消息（更新状态为删除，但不真正删除记录）
		msg.Status = 5 // 5-删除
		_, err = l.svcCtx.ImMessageModel.Update(ctx, session, msg)
		if err != nil {
			return errors.Wrapf(err, "update message status failed")
		}

		// 5. 如果这是最后一条消息，需要查找前一条消息更新会话信息
		// 获取该会话最新的未删除消息
		builder := l.svcCtx.ImMessageModel.SelectBuilder().
			Where("conversation_id = ? AND status != ?", msg.ConversationId, 5). // 排除已删除的消息
			OrderBy("id DESC").
			Limit(1)

		messages, err := l.svcCtx.ImMessageModel.FindAll(ctx, builder, "")
		if err == nil && len(messages) > 0 {
			latestMsg := messages[0]
			// 更新会话的最后消息信息（发送方和接收方的会话都要更新）
			senderConversationId := model.GenerateConversationId(msg.FromUserId, msg.ToUserId)
			receiverConversationId := model.GenerateConversationId(msg.ToUserId, msg.FromUserId)

			err = l.svcCtx.ImConversationModel.UpdateLastMessage(ctx, session, senderConversationId, latestMsg.Id, latestMsg.Content)
			if err != nil {
				l.Logger.Errorf("update sender conversation last message failed: %v", err)
			}

			err = l.svcCtx.ImConversationModel.UpdateLastMessage(ctx, session, receiverConversationId, latestMsg.Id, latestMsg.Content)
			if err != nil {
				l.Logger.Errorf("update receiver conversation last message failed: %v", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	l.Logger.Infof("Message deleted: userId=%d, messageId=%d", in.UserId, in.MessageId)

	return &message.DeleteMessageResp{
		Success: true,
	}, nil
}
