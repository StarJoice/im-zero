package logic

import (
	"context"
	"database/sql"
	"time"

	"im-zero/app/friend/cmd/rpc/friend"
	"im-zero/app/message/cmd/rpc/internal/svc"
	"im-zero/app/message/cmd/rpc/message"
	"im-zero/app/message/model"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SendMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMessageLogic {
	return &SendMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送消息
func (l *SendMessageLogic) SendMessage(in *message.SendMessageReq) (*message.SendMessageResp, error) {
	// 参数验证
	if in.FromUserId <= 0 || in.ToUserId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "from=%d, to=%d", in.FromUserId, in.ToUserId)
	}

	if in.MessageType <= 0 || in.MessageType > 5 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid message type"), "type=%d", in.MessageType)
	}

	if len(in.Content) == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "empty content"), "message content cannot be empty")
	}

	// 验证用户关系（检查是否为好友）
	_, err := l.svcCtx.FriendRpc.CheckFriendship(l.ctx, &friend.CheckFriendshipReq{
		UserId:       in.FromUserId,
		TargetUserId: in.ToUserId,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "check friendship failed")
	}

	var msgResp *message.SendMessageResp

	// 使用事务确保数据一致性
	err := l.svcCtx.ImMessageModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 查找或创建会话
		conversation, err := l.svcCtx.ImConversationModel.FindOrCreateConversation(ctx, session, in.FromUserId, in.ToUserId)
		if err != nil {
			return errors.Wrapf(err, "find or create conversation failed")
		}

		// 同时需要为接收方创建会话记录
		_, err = l.svcCtx.ImConversationModel.FindOrCreateConversation(ctx, session, in.ToUserId, in.FromUserId)
		if err != nil {
			return errors.Wrapf(err, "find or create receiver conversation failed")
		}

		// 2. 生成消息序号（简单的时间戳 + 用户ID）
		seq := time.Now().UnixNano()

		// 3. 创建消息记录
		msgModel := &model.ImMessage{
			FromUserId:     in.FromUserId,
			ToUserId:       in.ToUserId,
			ConversationId: conversation.ConversationId,
			MessageType:    int64(in.MessageType),
			Content:        in.Content,
			Extra:          sql.NullString{String: in.Extra, Valid: in.Extra != ""},
			Status:         1, // 1-已发送
			Seq:            seq,
		}

		result, err := l.svcCtx.ImMessageModel.Insert(ctx, session, msgModel)
		if err != nil {
			return errors.Wrapf(err, "insert message failed")
		}

		msgId, err := result.LastInsertId()
		if err != nil {
			return errors.Wrapf(err, "get message id failed")
		}
		msgModel.Id = msgId

		// 4. 更新发送方会话的最后消息信息
		err = l.svcCtx.ImConversationModel.UpdateLastMessage(ctx, session, conversation.ConversationId, msgId, in.Content)
		if err != nil {
			return errors.Wrapf(err, "update sender conversation failed")
		}

		// 5. 更新接收方会话的最后消息信息和未读计数
		receiverConversationId := model.GenerateConversationId(in.ToUserId, in.FromUserId)
		err = l.svcCtx.ImConversationModel.UpdateLastMessage(ctx, session, receiverConversationId, msgId, in.Content)
		if err != nil {
			return errors.Wrapf(err, "update receiver conversation failed")
		}

		err = l.svcCtx.ImConversationModel.IncrementUnreadCount(ctx, session, receiverConversationId, in.ToUserId)
		if err != nil {
			return errors.Wrapf(err, "increment unread count failed")
		}

		// 6. 构造返回结果
		msgResp = &message.SendMessageResp{
			Message: &message.MessageInfo{
				Id:             msgId,
				FromUserId:     msgModel.FromUserId,
				ToUserId:       msgModel.ToUserId,
				ConversationId: msgModel.ConversationId,
				MessageType:    int32(msgModel.MessageType),
				Content:        msgModel.Content,
				Extra:          msgModel.Extra.String,
				Status:         int32(msgModel.Status),
				CreateTime:     msgModel.CreateTime.Unix(),
				UpdateTime:     msgModel.UpdateTime.Unix(),
			},
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 推送消息给接收方
	pushLogic := NewPushMessageLogic(l.ctx, l.svcCtx)
	_, err = pushLogic.PushMessage(&message.PushMessageReq{
		UserId:  in.ToUserId,
		Message: msgResp.Message,
	})
	if err != nil {
		l.Logger.Errorf("Failed to push message to user %d: %v", in.ToUserId, err)
		// 推送失败不影响消息发送，只记录错误日志
	}

	l.Logger.Infof("Message sent successfully: from=%d, to=%d, msgId=%d", in.FromUserId, in.ToUserId, msgResp.Message.Id)

	return msgResp, nil
}
