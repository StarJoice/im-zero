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

type SendFriendRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendFriendRequestLogic {
	return &SendFriendRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送好友请求
func (l *SendFriendRequestLogic) SendFriendRequest(in *friend.SendFriendRequestReq) (*friend.SendFriendRequestResp, error) {
	// 参数验证
	if in.FromUserId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid from user id"), "fromUserId=%d", in.FromUserId)
	}
	if in.ToUserId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid to user id"), "toUserId=%d", in.ToUserId)
	}
	if in.FromUserId == in.ToUserId {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "cannot add yourself as friend"), "userId=%d", in.FromUserId)
	}
	if len(in.Message) > 200 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "message too long"), "message length=%d", len(in.Message))
	}

	// 检查是否已经是好友
	isFriend, err := l.svcCtx.ImFriendModel.CheckFriendship(l.ctx, in.FromUserId, in.ToUserId)
	if err != nil {
		return nil, errors.Wrapf(err, "check friendship failed")
	}
	if isFriend {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.FRIEND_ALREADY_EXISTS, "already friends"), "fromUserId=%d, toUserId=%d", in.FromUserId, in.ToUserId)
	}

	// 检查是否被拉黑
	isBlocked, hasBlocked, err := l.svcCtx.ImUserBlacklistModel.CheckMutualBlocked(l.ctx, in.FromUserId, in.ToUserId)
	if err != nil {
		return nil, errors.Wrapf(err, "check blocked status failed")
	}
	if isBlocked {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "you are blocked by target user"), "fromUserId=%d, toUserId=%d", in.FromUserId, in.ToUserId)
	}
	if hasBlocked {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PERMISSION_DENIED, "you have blocked target user"), "fromUserId=%d, toUserId=%d", in.FromUserId, in.ToUserId)
	}

	var requestId int64

	// 使用事务处理好友请求
	err = l.svcCtx.ImFriendRequestModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 检查是否有待处理的请求
		existingReq, err := l.svcCtx.ImFriendRequestModel.FindOneByFromToUserId(ctx, in.FromUserId, in.ToUserId)
		if err != nil && !errors.Is(err, model.ErrNotFound) {
			return errors.Wrapf(err, "check existing request failed")
		}

		if existingReq != nil {
			// 如果有待处理的请求，更新它
			if existingReq.Status == 0 {
				// 还在待处理状态，更新消息和过期时间
				existingReq.Message = sql.NullString{String: in.Message, Valid: in.Message != ""}
				existingReq.ExpireTime = sql.NullTime{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true} // 7天后过期
				_, err = l.svcCtx.ImFriendRequestModel.Update(ctx, session, existingReq)
				if err != nil {
					return errors.Wrapf(err, "update existing request failed")
				}
				requestId = existingReq.Id
				return nil
			}
			
			// 如果之前的请求已经被处理，检查间隔时间
			if existingReq.HandleTime.Valid && time.Since(existingReq.HandleTime.Time) < 24*time.Hour {
				return errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "please wait 24 hours before sending another request"), "lastRequestTime=%v", existingReq.HandleTime.Time)
			}
		}

		// 创建新的好友请求
		newRequest := &model.ImFriendRequest{
			FromUserId: in.FromUserId,
			ToUserId:   in.ToUserId,
			Message:    sql.NullString{String: in.Message, Valid: in.Message != ""},
			Status:     0, // 待处理
			ExpireTime: sql.NullTime{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true}, // 7天后过期
		}

		result, err := l.svcCtx.ImFriendRequestModel.Insert(ctx, session, newRequest)
		if err != nil {
			return errors.Wrapf(err, "insert friend request failed")
		}

		requestId, err = result.LastInsertId()
		if err != nil {
			return errors.Wrapf(err, "get request id failed")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &friend.SendFriendRequestResp{
		RequestId: requestId,
	}, nil
}
