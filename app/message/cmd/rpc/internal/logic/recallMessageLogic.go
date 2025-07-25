package logic

import (
	"context"
	"time"

	"im-zero/app/message/cmd/rpc/internal/svc"
	"im-zero/app/message/cmd/rpc/message"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RecallMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecallMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecallMessageLogic {
	return &RecallMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 撤回消息
func (l *RecallMessageLogic) RecallMessage(in *message.RecallMessageReq) (*message.RecallMessageResp, error) {
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

		// 2. 验证消息所有权
		if msg.FromUserId != in.UserId {
			return errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "permission denied"), "user %d cannot recall message from user %d", in.UserId, msg.FromUserId)
		}

		// 3. 检查消息状态（已撤回或已删除的消息不能再次操作）
		if msg.Status == 4 { // 4-撤回
			return errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "message already recalled"), "messageId=%d", in.MessageId)
		}

		if msg.Status == 5 { // 5-删除
			return errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "message already deleted"), "messageId=%d", in.MessageId)
		}

		// 4. 检查撤回时间限制（2分钟内可以撤回）
		if time.Since(msg.CreateTime) > 2*time.Minute {
			return errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "recall time limit exceeded"), "messageId=%d", in.MessageId)
		}

		// 5. 更新消息状态为撤回
		msg.Status = 4 // 4-撤回
		_, err = l.svcCtx.ImMessageModel.Update(ctx, session, msg)
		if err != nil {
			return errors.Wrapf(err, "update message status failed")
		}

		// 6. 如果这是最后一条消息，需要查找前一条消息更新会话信息
		lastMsg, err := l.svcCtx.ImMessageModel.FindLatestMessageByConversationId(ctx, msg.ConversationId)
		if err == nil && lastMsg.Id != msg.Id {
			// 找到了其他消息，更新会话的最后消息信息
			err = l.svcCtx.ImConversationModel.UpdateLastMessage(ctx, session, msg.ConversationId, lastMsg.Id, lastMsg.Content)
			if err != nil {
				l.Logger.Errorf("update conversation last message failed: %v", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	l.Logger.Infof("Message recalled: userId=%d, messageId=%d", in.UserId, in.MessageId)

	return &message.RecallMessageResp{
		Success: true,
	}, nil
}
