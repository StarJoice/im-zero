package logic

import (
	"context"

	"im-zero/app/message/cmd/rpc/internal/svc"
	"im-zero/app/message/cmd/rpc/message"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUnreadCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUnreadCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadCountLogic {
	return &GetUnreadCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取未读消息数
func (l *GetUnreadCountLogic) GetUnreadCount(in *message.GetUnreadCountReq) (*message.GetUnreadCountResp, error) {
	// 参数验证
	if in.UserId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", in.UserId)
	}

	var totalUnread int64
	
	if len(in.ConversationId) > 0 {
		// 获取指定会话的未读消息数
		count, err := l.svcCtx.ImMessageModel.CountUnreadMessages(l.ctx, in.UserId, in.ConversationId)
		if err != nil {
			return nil, errors.Wrapf(err, "count unread messages failed")
		}
		totalUnread = count
	} else {
		// 获取用户所有会话的未读消息总数
		// 通过会话表统计所有未读数
		conversations, _, err := l.svcCtx.ImConversationModel.FindConversationsByUserId(l.ctx, in.UserId, 1, 1000)
		if err != nil {
			return nil, errors.Wrapf(err, "find conversations failed")
		}
		
		for _, conv := range conversations {
			totalUnread += conv.UnreadCount
		}
	}

	return &message.GetUnreadCountResp{
		Count: totalUnread,
	}, nil
}
