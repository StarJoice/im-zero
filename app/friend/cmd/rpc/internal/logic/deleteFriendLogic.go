package logic

import (
	"context"

	"im-zero/app/friend/cmd/rpc/friend"
	"im-zero/app/friend/cmd/rpc/internal/svc"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DeleteFriendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFriendLogic {
	return &DeleteFriendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除好友
func (l *DeleteFriendLogic) DeleteFriend(in *friend.DeleteFriendReq) (*friend.DeleteFriendResp, error) {
	// 记录删除好友请求开始
	l.Logger.Infow("Start deleting friend",
		logx.Field("userId", in.UserId),
		logx.Field("friendUserId", in.FriendUserId),
	)

	// 参数验证
	if in.UserId <= 0 {
		l.Logger.Errorw("Invalid user id", logx.Field("userId", in.UserId))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", in.UserId)
	}
	if in.FriendUserId <= 0 {
		l.Logger.Errorw("Invalid friend user id", logx.Field("friendUserId", in.FriendUserId))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid friend user id"), "friendUserId=%d", in.FriendUserId)
	}
	if in.UserId == in.FriendUserId {
		l.Logger.Errorw("Cannot delete self as friend", logx.Field("userId", in.UserId))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "cannot delete self as friend"), "userId=%d", in.UserId)
	}

	// 检查是否为好友关系
	isFriend, err := l.svcCtx.ImFriendModel.CheckFriendship(l.ctx, in.UserId, in.FriendUserId)
	if err != nil {
		l.Logger.Errorw("Check friendship failed",
			logx.Field("userId", in.UserId),
			logx.Field("friendUserId", in.FriendUserId),
			logx.Field("error", err.Error()),
		)
		return nil, errors.Wrapf(err, "check friendship failed")
	}

	if !isFriend {
		l.Logger.Errorw("Users are not friends",
			logx.Field("userId", in.UserId),
			logx.Field("friendUserId", in.FriendUserId),
		)
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.NOT_FRIEND_RELATION, "users are not friends"), "userId=%d, friendUserId=%d", in.UserId, in.FriendUserId)
	}

	// 使用事务删除双向好友关系
	err = l.svcCtx.ImFriendModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 删除 userId -> friendUserId 的好友关系
		err := l.svcCtx.ImFriendModel.DeleteFriendship(ctx, session, in.UserId, in.FriendUserId)
		if err != nil {
			l.Logger.Errorw("Delete friendship failed",
				logx.Field("userId", in.UserId),
				logx.Field("friendUserId", in.FriendUserId),
				logx.Field("error", err.Error()),
			)
			return errors.Wrapf(err, "delete friendship failed")
		}

		// 删除 friendUserId -> userId 的好友关系
		err = l.svcCtx.ImFriendModel.DeleteFriendship(ctx, session, in.FriendUserId, in.UserId)
		if err != nil {
			l.Logger.Errorw("Delete reverse friendship failed",
				logx.Field("userId", in.UserId),
				logx.Field("friendUserId", in.FriendUserId),
				logx.Field("error", err.Error()),
			)
			return errors.Wrapf(err, "delete reverse friendship failed")
		}

		l.Logger.Infow("Successfully deleted mutual friendship",
			logx.Field("userId", in.UserId),
			logx.Field("friendUserId", in.FriendUserId),
		)

		return nil
	})

	if err != nil {
		l.Logger.Errorw("Delete friend transaction failed", logx.Field("error", err.Error()))
		return nil, err
	}

	l.Logger.Infow("Friend deleted successfully",
		logx.Field("userId", in.UserId),
		logx.Field("friendUserId", in.FriendUserId),
	)

	return &friend.DeleteFriendResp{
		Success: true,
	}, nil
}
