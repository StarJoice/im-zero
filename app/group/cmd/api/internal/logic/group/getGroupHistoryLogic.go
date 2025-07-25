package group

import (
	"context"

	"im-zero/app/group/cmd/api/internal/svc"
	"im-zero/app/group/cmd/api/internal/types"
	"im-zero/app/group/cmd/rpc/group"
	"im-zero/pkg/ctxdata"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取群聊记录
func NewGetGroupHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupHistoryLogic {
	return &GetGroupHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGroupHistoryLogic) GetGroupHistory(req *types.GetGroupHistoryReq) (resp *types.GetGroupHistoryResp, err error) {
	// 参数验证
	if req.GroupId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid group id"), "groupId=%d", req.GroupId)
	}

	// 获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.UNAUTHORIZED, "user not login"), "userId is empty")
	}

	// 设置默认值
	if req.Limit <= 0 {
		req.Limit = 20
	}

	// 调用RPC服务获取群消息历史
	rpcResp, err := l.svcCtx.GroupRpc.GetGroupHistory(l.ctx, &group.GetGroupHistoryReq{
		GroupId:       req.GroupId,
		UserId:        userId,
		LastMessageId: req.LastMessageId,
		Limit:         req.Limit,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get group history from rpc failed")
	}

	// 转换返回结果
	messages := make([]types.GroupMessage, 0, len(rpcResp.Messages))
	for _, msg := range rpcResp.Messages {
		messages = append(messages, types.GroupMessage{
			Id:          msg.Id,
			GroupId:     msg.GroupId,
			FromUserId:  msg.FromUserId,
			MessageType: msg.MessageType,
			Content:     msg.Content,
			Extra:       msg.Extra,
			Status:      msg.Status,
			CreateTime:  msg.CreateTime,
		})
	}

	return &types.GetGroupHistoryResp{
		Messages: messages,
		HasMore:  rpcResp.HasMore,
	}, nil
}
