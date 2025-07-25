package logic

import (
	"context"
	"database/sql"
	"time"

	"im-zero/app/friend/cmd/rpc/friend"
	"im-zero/app/friend/cmd/rpc/internal/svc"
	"im-zero/app/friend/model"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type HandleFriendRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleFriendRequestLogic {
	return &HandleFriendRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 处理好友请求
func (l *HandleFriendRequestLogic) HandleFriendRequest(in *friend.HandleFriendRequestReq) (*friend.HandleFriendRequestResp, error) {
	// 记录处理开始
	l.Logger.Infow("Start handling friend request",
		logx.Field("requestId", in.RequestId),
		logx.Field("userId", in.UserId),
		logx.Field("action", in.Action),
		logx.Field("remark", in.Remark),
	)

	// 参数验证
	if in.RequestId <= 0 {
		l.Logger.Errorw("Invalid request id", logx.Field("requestId", in.RequestId))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid request id"), "requestId=%d", in.RequestId)
	}
	if in.UserId <= 0 {
		l.Logger.Errorw("Invalid user id", logx.Field("userId", in.UserId))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", in.UserId)
	}
	if in.Action != 1 && in.Action != 2 {
		l.Logger.Errorw("Invalid action", logx.Field("action", in.Action))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid action"), "action=%d", in.Action)
	}
	if len(in.Remark) > 50 {
		l.Logger.Errorw("Remark too long", logx.Field("remarkLength", len(in.Remark)))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "remark too long"), "remark length=%d", len(in.Remark))
	}

	// 获取好友请求信息
	request, err := l.svcCtx.ImFriendRequestModel.FindOne(l.ctx, in.RequestId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorw("Friend request not found", logx.Field("requestId", in.RequestId))
			return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.FRIEND_REQUEST_NOT_FOUND, "friend request not found"), "requestId=%d", in.RequestId)
		}
		l.Logger.Errorw("Find friend request failed", logx.Field("requestId", in.RequestId), logx.Field("error", err.Error()))
		return nil, errors.Wrapf(err, "find friend request failed")
	}

	l.Logger.Infow("Found friend request",
		logx.Field("requestId", request.Id),
		logx.Field("fromUserId", request.FromUserId),
		logx.Field("toUserId", request.ToUserId),
		logx.Field("status", request.Status),
	)

	// 检查权限（只有请求的接收者可以处理）
	if request.ToUserId != in.UserId {
		l.Logger.Errorw("Permission denied: only request receiver can handle",
			logx.Field("requestId", in.RequestId),
			logx.Field("userId", in.UserId),
			logx.Field("toUserId", request.ToUserId),
		)
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "only request receiver can handle"), "requestId=%d, userId=%d, toUserId=%d", in.RequestId, in.UserId, request.ToUserId)
	}

	// 检查请求状态
	if request.Status != 0 {
		l.Logger.Errorw("Request already handled",
			logx.Field("requestId", in.RequestId),
			logx.Field("status", request.Status),
		)
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.FRIEND_REQUEST_ALREADY_HANDLED, "request already handled"), "requestId=%d, status=%d", in.RequestId, request.Status)
	}

	// 检查是否过期
	if request.ExpireTime.Valid && request.ExpireTime.Time.Before(time.Now()) {
		l.Logger.Errorw("Request has expired",
			logx.Field("requestId", in.RequestId),
			logx.Field("expireTime", request.ExpireTime.Time),
		)
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.FRIEND_REQUEST_EXPIRED, "request has expired"), "requestId=%d, expireTime=%v", in.RequestId, request.ExpireTime.Time)
	}

	// 使用事务处理好友请求
	err = l.svcCtx.ImFriendRequestModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 更新请求状态
		request.Status = int64(in.Action) // 1-同意，2-拒绝
		request.HandleTime = sql.NullTime{Time: time.Now(), Valid: true}
		_, err = l.svcCtx.ImFriendRequestModel.Update(ctx, session, request)
		if err != nil {
			l.Logger.Errorw("Update request status failed", logx.Field("error", err.Error()))
			return errors.Wrapf(err, "update request status failed")
		}

		l.Logger.Infow("Updated friend request status",
			logx.Field("requestId", in.RequestId),
			logx.Field("action", in.Action),
		)

		// 如果同意请求，创建好友关系
		if in.Action == 1 {
			l.Logger.Infow("Creating friend relationship",
				logx.Field("fromUserId", request.FromUserId),
				logx.Field("toUserId", request.ToUserId),
			)

			// 检查是否已经是好友
			isFriend, err := l.svcCtx.ImFriendModel.CheckFriendship(ctx, request.FromUserId, request.ToUserId)
			if err != nil {
				l.Logger.Errorw("Check friendship failed", logx.Field("error", err.Error()))
				return errors.Wrapf(err, "check friendship failed")
			}
			if isFriend {
				l.Logger.Warnw("Users are already friends",
					logx.Field("fromUserId", request.FromUserId),
					logx.Field("toUserId", request.ToUserId),
				)
				return errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.FRIEND_ALREADY_EXISTS, "already friends"), "fromUserId=%d, toUserId=%d", request.FromUserId, request.ToUserId)
			}

			// 检查拉黑状态
			isBlocked, hasBlocked, err := l.svcCtx.ImUserBlacklistModel.CheckMutualBlocked(ctx, request.FromUserId, request.ToUserId)
			if err != nil {
				l.Logger.Errorw("Check blocked status failed", logx.Field("error", err.Error()))
				return errors.Wrapf(err, "check blocked status failed")
			}
			if isBlocked || hasBlocked {
				l.Logger.Errorw("Users are in block relationship",
					logx.Field("fromUserId", request.FromUserId),
					logx.Field("toUserId", request.ToUserId),
					logx.Field("isBlocked", isBlocked),
					logx.Field("hasBlocked", hasBlocked),
				)
				return errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.USER_BLOCKED, "blocked users cannot be friends"), "fromUserId=%d, toUserId=%d", request.FromUserId, request.ToUserId)
			}

			// 创建双向好友关系
			now := time.Now()

			// 为发起者创建好友记录
			fromFriend := &model.ImFriend{
				UserId:   request.FromUserId,
				FriendId: request.ToUserId,
				Remark:   sql.NullString{String: "", Valid: false}, // 发起者暂时没有备注
				Status:   1,                                        // 正常状态
			}
			fromFriend.CreateTime = now
			fromFriend.UpdateTime = now

			_, err = l.svcCtx.ImFriendModel.Insert(ctx, session, fromFriend)
			if err != nil {
				l.Logger.Errorw("Insert from friend failed", logx.Field("error", err.Error()))
				return errors.Wrapf(err, "insert from friend failed")
			}

			// 为接收者创建好友记录
			toFriend := &model.ImFriend{
				UserId:   request.ToUserId,
				FriendId: request.FromUserId,
				Remark:   sql.NullString{String: in.Remark, Valid: in.Remark != ""}, // 接收者可以设置备注
				Status:   1,                                                          // 正常状态
			}
			toFriend.CreateTime = now
			toFriend.UpdateTime = now

			_, err = l.svcCtx.ImFriendModel.Insert(ctx, session, toFriend)
			if err != nil {
				l.Logger.Errorw("Insert to friend failed", logx.Field("error", err.Error()))
				return errors.Wrapf(err, "insert to friend failed")
			}

			l.Logger.Infow("Successfully created friend relationship",
				logx.Field("fromUserId", request.FromUserId),
				logx.Field("toUserId", request.ToUserId),
				logx.Field("remark", in.Remark),
			)
		}

		return nil
	})

	if err != nil {
		l.Logger.Errorw("Handle friend request transaction failed", logx.Field("error", err.Error()))
		return nil, err
	}

	actionStr := "rejected"
	if in.Action == 1 {
		actionStr = "accepted"
	}

	l.Logger.Infow("Friend request handled successfully",
		logx.Field("requestId", in.RequestId),
		logx.Field("userId", in.UserId),
		logx.Field("action", actionStr),
		logx.Field("fromUserId", request.FromUserId),
		logx.Field("toUserId", request.ToUserId),
	)

	return &friend.HandleFriendRequestResp{
		Success: true,
	}, nil
}
