package logic

import (
	"context"

	"im-zero/app/message/cmd/rpc/internal/svc"
	"im-zero/app/message/cmd/rpc/message"
	"im-zero/app/usercenter/cmd/rpc/usercenter"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetConversationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetConversationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationsLogic {
	return &GetConversationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取对话列表
func (l *GetConversationsLogic) GetConversations(in *message.GetConversationsReq) (*message.GetConversationsResp, error) {
	// 参数验证
	if in.UserId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", in.UserId)
	}

	page := in.Page
	if page <= 0 {
		page = 1
	}

	limit := in.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	// 查询用户的对话列表
	conversations, total, err := l.svcCtx.ImConversationModel.FindConversationsByUserId(l.ctx, in.UserId, int64(page), int64(limit))
	if err != nil {
		return nil, errors.Wrapf(err, "find conversations failed")
	}

	var conversationInfos []*message.ConversationInfo
	for _, conv := range conversations {
		// 获取对方用户信息
		friendInfo, err := l.svcCtx.UsercenterRpc.GetUserInfo(l.ctx, &usercenter.GetUserInfoReq{
			Id: conv.FriendId,
		})
		if err != nil {
			l.Logger.Errorf("get friend info failed: friendId=%d, err=%v", conv.FriendId, err)
			continue // 跳过这个会话，继续处理其他的
		}

		// 构建最后一条消息信息
		var lastMessage *message.MessageInfo
		if conv.LastMessageId.Valid && conv.LastMessageId.Int64 > 0 {
			lastMessage = &message.MessageInfo{
				Id:         conv.LastMessageId.Int64,
				Content:    conv.LastMessageContent.String,
				CreateTime: conv.LastMessageTime.Time.Unix(),
			}
		}

		conversationInfo := &message.ConversationInfo{
			Id:          conv.ConversationId,
			Type:        int32(conv.ConversationType), // 1:单聊 2:群聊
			Name:        friendInfo.User.Nickname,
			Avatar:      friendInfo.User.Avatar,
			LastMessage: lastMessage,
			UnreadCount: int32(conv.UnreadCount),
			UpdateTime:  conv.UpdateTime.Unix(),
		}

		conversationInfos = append(conversationInfos, conversationInfo)
	}

	return &message.GetConversationsResp{
		Conversations: conversationInfos,
		Total:         int32(total),
	}, nil
}
