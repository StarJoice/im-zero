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

type BlockUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBlockUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlockUserLogic {
	return &BlockUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 拉黑用户
func (l *BlockUserLogic) BlockUser(in *friend.BlockUserReq) (*friend.BlockUserResp, error) {
	// 记录拉黑用户请求开始
	l.Logger.Infow("Start blocking user",
		logx.Field("userId", in.UserId),
		logx.Field("blockedUserId", in.BlockedUserId),
		logx.Field("reason", in.Reason),
	)

	// 参数验证
	if in.UserId <= 0 {
		l.Logger.Errorw("Invalid user id", logx.Field("userId", in.UserId))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", in.UserId)
	}
	if in.BlockedUserId <= 0 {
		l.Logger.Errorw("Invalid blocked user id", logx.Field("blockedUserId", in.BlockedUserId))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid blocked user id"), "blockedUserId=%d", in.BlockedUserId)
	}
	if in.UserId == in.BlockedUserId {
		l.Logger.Errorw("Cannot block self", logx.Field("userId", in.UserId))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "cannot block self"), "userId=%d", in.UserId)
	}
	if len(in.Reason) > 100 {
		l.Logger.Errorw("Block reason too long", logx.Field("reasonLength", len(in.Reason)))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "block reason too long"), "reason length=%d", len(in.Reason))
	}

	// 检查是否已经拉黑
	isBlocked, err := l.svcCtx.ImUserBlacklistModel.CheckBlocked(l.ctx, in.UserId, in.BlockedUserId)
	if err != nil {
		l.Logger.Errorw("Check blocked status failed",
			logx.Field("userId", in.UserId),
			logx.Field("blockedUserId", in.BlockedUserId),
			logx.Field("error", err.Error()),
		)
		return nil, errors.Wrapf(err, "check blocked status failed")
	}

	if isBlocked {
		l.Logger.Errorw("User already blocked",
			logx.Field("userId", in.UserId),
			logx.Field("blockedUserId", in.BlockedUserId),
		)
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.USER_ALREADY_BLOCKED, "user already blocked"), "userId=%d, blockedUserId=%d", in.UserId, in.BlockedUserId)
	}

	// 使用事务处理拉黑操作
	err = l.svcCtx.ImUserBlacklistModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 如果是好友关系，先删除好友关系
		isFriend, err := l.svcCtx.ImFriendModel.CheckFriendship(ctx, in.UserId, in.BlockedUserId)
		if err != nil {
			l.Logger.Errorw("Check friendship failed", logx.Field("error", err.Error()))
			return errors.Wrapf(err, "check friendship failed")
		}

		if isFriend {
			l.Logger.Infow("Deleting friend relationship before blocking",
				logx.Field("userId", in.UserId),
				logx.Field("blockedUserId", in.BlockedUserId),
			)

			// 删除双向好友关系
			err = l.svcCtx.ImFriendModel.DeleteFriendship(ctx, session, in.UserId, in.BlockedUserId)
			if err != nil {
				l.Logger.Errorw("Delete friendship failed", logx.Field("error", err.Error()))
				return errors.Wrapf(err, "delete friendship failed")
			}

			err = l.svcCtx.ImFriendModel.DeleteFriendship(ctx, session, in.BlockedUserId, in.UserId)
			if err != nil {
				l.Logger.Errorw("Delete reverse friendship failed", logx.Field("error", err.Error()))
				return errors.Wrapf(err, "delete reverse friendship failed")
			}
		}

		// 创建拉黑记录
		now := time.Now()
		blacklist := &model.ImUserBlacklist{
			UserId:        in.UserId,
			BlockedUserId: in.BlockedUserId,
			Reason:        sql.NullString{String: in.Reason, Valid: in.Reason != ""},
		}
		blacklist.CreateTime = now
		blacklist.UpdateTime = now

		_, err = l.svcCtx.ImUserBlacklistModel.Insert(ctx, session, blacklist)
		if err != nil {
			l.Logger.Errorw("Insert blacklist record failed", logx.Field("error", err.Error()))
			return errors.Wrapf(err, "insert blacklist record failed")
		}

		l.Logger.Infow("Successfully blocked user",
			logx.Field("userId", in.UserId),
			logx.Field("blockedUserId", in.BlockedUserId),
			logx.Field("reason", in.Reason),
		)

		return nil
	})

	if err != nil {
		l.Logger.Errorw("Block user transaction failed", logx.Field("error", err.Error()))
		return nil, err
	}

	l.Logger.Infow("User blocked successfully",
		logx.Field("userId", in.UserId),
		logx.Field("blockedUserId", in.BlockedUserId),
	)

	return &friend.BlockUserResp{
		Success: true,
	}, nil
}
