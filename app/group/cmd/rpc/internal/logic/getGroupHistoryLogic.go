package logic

import (
	"context"

	"im-zero/app/group/cmd/rpc/group"
	"im-zero/app/group/cmd/rpc/internal/svc"
	"im-zero/app/group/model"
	"im-zero/pkg/xerrs"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupHistoryLogic {
	return &GetGroupHistoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取群聊记录
func (l *GetGroupHistoryLogic) GetGroupHistory(in *group.GetGroupHistoryReq) (*group.GetGroupHistoryResp, error) {
	// 参数验证
	if in.GroupId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid group id"), "groupId=%d", in.GroupId)
	}
	if in.UserId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", in.UserId)
	}

	// 设置默认值
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// 检查群组是否存在
	groupInfo, err := l.svcCtx.ImGroupModel.FindOne(l.ctx, in.GroupId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.GROUP_NOT_FOUND, "group not found"), "groupId=%d", in.GroupId)
		}
		return nil, errors.Wrapf(err, "find group failed")
	}

	// 检查群组状态
	if groupInfo.Status != 1 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.GROUP_ALREADY_DISSOLVED, "group is not active"), "groupId=%d, status=%d", in.GroupId, groupInfo.Status)
	}

	// 检查用户是否在群中
	member, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, in.UserId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.GROUP_MEMBER_NOT_FOUND, "user not in group"), "userId=%d, groupId=%d", in.UserId, in.GroupId)
		}
		return nil, errors.Wrapf(err, "check group member failed")
	}

	// 检查用户状态
	if member.Status != 1 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.GROUP_MEMBER_NOT_FOUND, "user not active in group"), "userId=%d, status=%d", in.UserId, member.Status)
	}

	// 构建查询条件
	messageBuilder := l.svcCtx.ImGroupMessageModel.SelectBuilder().
		Where(squirrel.Eq{"group_id": in.GroupId}).
		Where(squirrel.NotEq{"status": 5}) // 排除已删除的消息

	// 如果指定了起始消息ID，则查询该消息之前的消息
	if in.LastMessageId > 0 {
		messageBuilder = messageBuilder.Where(squirrel.Lt{"id": in.LastMessageId})
	}

	// 按消息ID倒序排列，获取最新的消息
	messages, err := l.svcCtx.ImGroupMessageModel.FindPageListByPage(
		l.ctx,
		messageBuilder,
		1, // 固定为第一页
		int64(limit),
		"id DESC", // 按ID倒序
	)
	if err != nil {
		return nil, errors.Wrapf(err, "find group messages failed")
	}

	// 转换为响应格式
	var messageInfos []*group.GroupMessageInfo
	for _, msg := range messages {
		messageInfo := &group.GroupMessageInfo{
			Id:          msg.Id,
			GroupId:     msg.GroupId,
			FromUserId:  msg.FromUserId,
			MessageType: int32(msg.MessageType),
			Content:     msg.Content,
			Extra:       msg.Extra.String,
			Status:      int32(msg.Status),
			CreateTime:  msg.CreateTime.Unix(),
		}
		messageInfos = append(messageInfos, messageInfo)
	}

	// 判断是否还有更多消息
	hasMore := len(messages) == int(limit)
	if hasMore && len(messages) > 0 {
		// 检查是否还有更早的消息
		oldestMsgId := messages[len(messages)-1].Id
		checkBuilder := l.svcCtx.ImGroupMessageModel.SelectBuilder().
			Where(squirrel.Eq{"group_id": in.GroupId}).
			Where(squirrel.Lt{"id": oldestMsgId}).
			Where(squirrel.NotEq{"status": 5})

		count, err := l.svcCtx.ImGroupMessageModel.FindCount(l.ctx, checkBuilder, "id")
		if err != nil {
			l.Logger.Errorf("check has more messages failed: %v", err)
			hasMore = false
		} else {
			hasMore = count > 0
		}
	}

	return &group.GetGroupHistoryResp{
		Messages: messageInfos,
		HasMore:  hasMore,
	}, nil
}
