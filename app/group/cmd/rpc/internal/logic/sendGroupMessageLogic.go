package logic

import (
	"context"
	"database/sql"
	"time"

	"im-zero/app/group/cmd/rpc/group"
	"im-zero/app/group/cmd/rpc/internal/svc"
	"im-zero/app/group/model"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SendGroupMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendGroupMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendGroupMessageLogic {
	return &SendGroupMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送群消息
func (l *SendGroupMessageLogic) SendGroupMessage(in *group.SendGroupMessageReq) (*group.SendGroupMessageResp, error) {
	// 参数验证
	if in.GroupId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid group id"), "groupId=%d", in.GroupId)
	}
	if in.FromUserId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid from user id"), "fromUserId=%d", in.FromUserId)
	}
	if len(in.Content) == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "message content is required"), "content is empty")
	}

	// 检查用户是否在群中且有发言权限
	member, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, in.FromUserId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.NOT_FOUND, "user not in group"), "userId=%d, groupId=%d", in.FromUserId, in.GroupId)
		}
		return nil, errors.Wrapf(err, "check group member failed")
	}

	// 检查用户状态
	if member.Status != 1 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "user not active in group"), "userId=%d, status=%d", in.FromUserId, member.Status)
	}

	// 检查是否被禁言
	if member.MuteEndTime.Valid && member.MuteEndTime.Time.After(time.Now()) {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "user is muted"), "userId=%d", in.FromUserId)
	}

	// 获取群组信息
	groupInfo, err := l.svcCtx.ImGroupModel.FindOne(l.ctx, in.GroupId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.NOT_FOUND, "group not found"), "groupId=%d", in.GroupId)
		}
		return nil, errors.Wrapf(err, "find group failed")
	}

	// 检查群组状态
	if groupInfo.Status != 1 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "group is not active"), "groupId=%d, status=%d", in.GroupId, groupInfo.Status)
	}

	// 创建群消息
	var messageInfo *model.ImGroupMessage
	err = l.svcCtx.ImGroupMessageModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 生成消息序号（使用群组version字段确保唯一性和顺序性）
		seq, err := l.svcCtx.ImGroupModel.GetNextMessageSeq(ctx, session, in.GroupId)
		if err != nil {
			return errors.Wrapf(err, "get next message seq failed")
		}

		newMessage := &model.ImGroupMessage{
			GroupId:     in.GroupId,
			FromUserId:  in.FromUserId,
			MessageType: int64(in.MessageType),
			Content:     in.Content,
			Extra:       sql.NullString{String: in.Extra, Valid: in.Extra != ""},
			Status:      1, // 已发送
			Seq:         seq,
		}

		_, err = l.svcCtx.ImGroupMessageModel.Insert(ctx, session, newMessage)
		if err != nil {
			return errors.Wrapf(err, "insert group message failed")
		}

		messageInfo = newMessage
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 返回消息信息
	return &group.SendGroupMessageResp{
		Message: &group.GroupMessageInfo{
			Id:          messageInfo.Id,
			GroupId:     messageInfo.GroupId,
			FromUserId:  messageInfo.FromUserId,
			MessageType: int32(messageInfo.MessageType),
			Content:     messageInfo.Content,
			Extra:       messageInfo.Extra.String,
			Status:      int32(messageInfo.Status),
			CreateTime:  messageInfo.CreateTime.Unix(),
		},
	}, nil
}
