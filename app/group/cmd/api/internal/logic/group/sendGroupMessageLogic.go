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

type SendGroupMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 群组消息发送
func NewSendGroupMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendGroupMessageLogic {
	return &SendGroupMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendGroupMessageLogic) SendGroupMessage(req *types.SendGroupMessageReq) (resp *types.SendGroupMessageResp, err error) {
	// 参数验证
	if req.GroupId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid group id"), "groupId=%d", req.GroupId)
	}
	if req.MessageType <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid message type"), "messageType=%d", req.MessageType)
	}
	if len(req.Content) == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "content is required"), "content is empty")
	}

	// 获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.UNAUTHORIZED, "user not login"), "userId is empty")
	}

	// 调用RPC服务发送群消息
	rpcResp, err := l.svcCtx.GroupRpc.SendGroupMessage(l.ctx, &group.SendGroupMessageReq{
		GroupId:     req.GroupId,
		FromUserId:  userId,
		MessageType: req.MessageType,
		Content:     req.Content,
		Extra:       req.Extra,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "send group message from rpc failed")
	}

	// 转换返回结果
	return &types.SendGroupMessageResp{
		Message: types.GroupMessage{
			Id:          rpcResp.Message.Id,
			GroupId:     rpcResp.Message.GroupId,
			FromUserId:  rpcResp.Message.FromUserId,
			MessageType: rpcResp.Message.MessageType,
			Content:     rpcResp.Message.Content,
			Extra:       rpcResp.Message.Extra,
			Status:      rpcResp.Message.Status,
			CreateTime:  rpcResp.Message.CreateTime,
		},
	}, nil
}
