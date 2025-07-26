package friend

import (
	"context"

	"im-zero/app/friend/cmd/api/internal/svc"
	"im-zero/app/friend/cmd/api/internal/types"
	"im-zero/app/friend/cmd/rpc/friend"
	"im-zero/pkg/ctxdata"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetFriendRequestsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取好友请求列表
func NewGetFriendRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendRequestsLogic {
	return &GetFriendRequestsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFriendRequestsLogic) GetFriendRequests(req *types.GetFriendRequestsReq) (resp *types.GetFriendRequestsResp, err error) {
	// 参数验证
	if req.Type != 1 && req.Type != 2 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid type"), "type must be 1(sent) or 2(received), got=%d", req.Type)
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Limit > 50 {
		req.Limit = 50
	}

	// 获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.UNAUTHORIZED, "user not login"), "userId is empty")
	}

	// 调用RPC服务获取好友请求列表
	rpcResp, err := l.svcCtx.FriendRpc.GetFriendRequests(l.ctx, &friend.GetFriendRequestsReq{
		UserId: userId,
		Type:   req.Type,
		Page:   req.Page,
		Limit:  req.Limit,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get friend requests rpc failed")
	}

	// 转换返回结果
	var requests []types.FriendRequest
	for _, requestInfo := range rpcResp.Requests {
		requests = append(requests, types.FriendRequest{
			Id:      requestInfo.Id,
			Message: requestInfo.Message,
			Status:  requestInfo.Status,
			FromUser: types.UserInfo{
				Id:       requestInfo.FromUser.Id,
				Mobile:   requestInfo.FromUser.Mobile,
				Nickname: requestInfo.FromUser.Nickname,
				Avatar:   requestInfo.FromUser.Avatar,
				Sign:     requestInfo.FromUser.Sign,
				Status:   requestInfo.FromUser.Status,
			},
			ToUser: types.UserInfo{
				Id:       requestInfo.ToUser.Id,
				Mobile:   requestInfo.ToUser.Mobile,
				Nickname: requestInfo.ToUser.Nickname,
				Avatar:   requestInfo.ToUser.Avatar,
				Sign:     requestInfo.ToUser.Sign,
				Status:   requestInfo.ToUser.Status,
			},
			CreateTime: requestInfo.CreateTime,
			HandleTime: requestInfo.HandleTime,
		})
	}

	// 记录操作日志
	l.Logger.Infof("Get friend requests successfully: userId=%d, type=%d, count=%d", userId, req.Type, len(requests))

	return &types.GetFriendRequestsResp{
		Requests: requests,
		Total:    rpcResp.Total,
	}, nil
}
