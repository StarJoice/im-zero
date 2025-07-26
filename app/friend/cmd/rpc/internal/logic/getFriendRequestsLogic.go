package logic

import (
	"context"

	"im-zero/app/friend/cmd/rpc/friend"
	"im-zero/app/friend/cmd/rpc/internal/svc"
	"im-zero/app/friend/model"
	"im-zero/app/usercenter/cmd/rpc/usercenter"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetFriendRequestsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFriendRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendRequestsLogic {
	return &GetFriendRequestsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取好友请求列表
func (l *GetFriendRequestsLogic) GetFriendRequests(in *friend.GetFriendRequestsReq) (*friend.GetFriendRequestsResp, error) {
	// 记录获取好友请求列表请求开始
	l.Logger.Infow("Start getting friend requests",
		logx.Field("userId", in.UserId),
		logx.Field("type", in.Type),
		logx.Field("page", in.Page),
		logx.Field("limit", in.Limit),
	)

	// 参数验证
	if in.UserId <= 0 {
		l.Logger.Errorw("Invalid user id", logx.Field("userId", in.UserId))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", in.UserId)
	}
	if in.Type < 0 || in.Type > 2 {
		l.Logger.Errorw("Invalid type", logx.Field("type", in.Type))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid type"), "type=%d", in.Type)
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.Limit <= 0 || in.Limit > 50 {
		in.Limit = 20
	}

	// 根据类型获取好友请求列表
	var requests []*friend.FriendRequestInfo
	var total int32

	switch in.Type {
	case 0: // 全部请求
		// 获取我发出的请求
		sentRequests, err := l.svcCtx.ImFriendRequestModel.FindSentRequests(l.ctx, in.UserId, in.Page, in.Limit)
		if err != nil {
			l.Logger.Errorw("Find sent requests failed", logx.Field("error", err.Error()))
			return nil, errors.Wrapf(err, "find sent requests failed")
		}

		// 获取我收到的请求
		receivedRequests, err := l.svcCtx.ImFriendRequestModel.FindReceivedRequests(l.ctx, in.UserId, in.Page, in.Limit)
		if err != nil {
			l.Logger.Errorw("Find received requests failed", logx.Field("error", err.Error()))
			return nil, errors.Wrapf(err, "find received requests failed")
		}

		// 合并结果
		allRequests := append(sentRequests, receivedRequests...)
		total = int32(len(allRequests))

		// 转换结果
		for _, req := range allRequests {
			reqInfo, err := l.convertToFriendRequestInfo(req)
			if err != nil {
				l.Logger.Errorw("Convert request info failed", logx.Field("requestId", req.Id), logx.Field("error", err.Error()))
				continue
			}
			requests = append(requests, reqInfo)
		}

	case 1: // 我发出的请求
		sentRequests, err := l.svcCtx.ImFriendRequestModel.FindSentRequests(l.ctx, in.UserId, in.Page, in.Limit)
		if err != nil {
			l.Logger.Errorw("Find sent requests failed", logx.Field("error", err.Error()))
			return nil, errors.Wrapf(err, "find sent requests failed")
		}

		total = int32(len(sentRequests))
		for _, req := range sentRequests {
			reqInfo, err := l.convertToFriendRequestInfo(req)
			if err != nil {
				l.Logger.Errorw("Convert request info failed", logx.Field("requestId", req.Id), logx.Field("error", err.Error()))
				continue
			}
			requests = append(requests, reqInfo)
		}

	case 2: // 我收到的请求
		receivedRequests, err := l.svcCtx.ImFriendRequestModel.FindReceivedRequests(l.ctx, in.UserId, in.Page, in.Limit)
		if err != nil {
			l.Logger.Errorw("Find received requests failed", logx.Field("error", err.Error()))
			return nil, errors.Wrapf(err, "find received requests failed")
		}

		total = int32(len(receivedRequests))
		for _, req := range receivedRequests {
			reqInfo, err := l.convertToFriendRequestInfo(req)
			if err != nil {
				l.Logger.Errorw("Convert request info failed", logx.Field("requestId", req.Id), logx.Field("error", err.Error()))
				continue
			}
			requests = append(requests, reqInfo)
		}
	}

	l.Logger.Infow("Get friend requests successfully",
		logx.Field("userId", in.UserId),
		logx.Field("type", in.Type),
		logx.Field("total", total),
		logx.Field("returnedCount", len(requests)),
	)

	return &friend.GetFriendRequestsResp{
		Requests: requests,
		Total:    total,
	}, nil
}

// 转换为 FriendRequestInfo
func (l *GetFriendRequestsLogic) convertToFriendRequestInfo(req *model.ImFriendRequest) (*friend.FriendRequestInfo, error) {
	// 获取发起者用户信息
	fromUserInfo, err := l.svcCtx.UsercenterRpc.GetUserInfo(l.ctx, &usercenter.GetUserInfoReq{
		Id: req.FromUserId,
	})
	if err != nil {
		l.Logger.Errorw("Get from user info failed",
			logx.Field("fromUserId", req.FromUserId),
			logx.Field("error", err.Error()),
		)
		return nil, errors.Wrapf(err, "get from user info failed")
	}

	// 获取接收者用户信息
	toUserInfo, err := l.svcCtx.UsercenterRpc.GetUserInfo(l.ctx, &usercenter.GetUserInfoReq{
		Id: req.ToUserId,
	})
	if err != nil {
		l.Logger.Errorw("Get to user info failed",
			logx.Field("toUserId", req.ToUserId),
			logx.Field("error", err.Error()),
		)
		return nil, errors.Wrapf(err, "get to user info failed")
	}

	// 构建响应
	friendRequestInfo := &friend.FriendRequestInfo{
		Id: req.Id,
		FromUser: &friend.UserInfo{
			Id:       fromUserInfo.User.Id,
			Mobile:   fromUserInfo.User.Mobile,
			Nickname: fromUserInfo.User.Nickname,
			Avatar:   fromUserInfo.User.Avatar,
			Sign:     fromUserInfo.User.Sign,
		},
		ToUser: &friend.UserInfo{
			Id:       toUserInfo.User.Id,
			Mobile:   toUserInfo.User.Mobile,
			Nickname: toUserInfo.User.Nickname,
			Avatar:   toUserInfo.User.Avatar,
			Sign:     toUserInfo.User.Sign,
		},
		Message:    req.Message.String,
		Status:     int32(req.Status),
		CreateTime: req.CreateTime.Unix(),
	}

	// 设置处理时间（如果有）
	if req.HandleTime.Valid {
		friendRequestInfo.HandleTime = req.HandleTime.Time.Unix()
	}

	return friendRequestInfo, nil
}
