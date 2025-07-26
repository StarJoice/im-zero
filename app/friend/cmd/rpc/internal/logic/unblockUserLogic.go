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

type UnblockUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnblockUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnblockUserLogic {
	return &UnblockUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 取消拉黑
func (l *UnblockUserLogic) UnblockUser(in *friend.UnblockUserReq) (*friend.UnblockUserResp, error) {
	// 记录取消拉黑请求开始
	l.Logger.Infow("Start unblocking user",
		logx.Field("userId", in.UserId),
		logx.Field("blockedUserId", in.BlockedUserId),
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
		l.Logger.Errorw("Cannot unblock self", logx.Field("userId", in.UserId))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "cannot unblock self"), "userId=%d", in.UserId)
	}

	// 检查是否已拉黑
	isBlocked, err := l.svcCtx.ImUserBlacklistModel.CheckBlocked(l.ctx, in.UserId, in.BlockedUserId)
	if err != nil {
		l.Logger.Errorw("Check blocked status failed",
			logx.Field("userId", in.UserId),
			logx.Field("blockedUserId", in.BlockedUserId),
			logx.Field("error", err.Error()),
		)
		return nil, errors.Wrapf(err, "check blocked status failed")
	}

	if !isBlocked {
		l.Logger.Errorw("User not blocked",
			logx.Field("userId", in.UserId),
			logx.Field("blockedUserId", in.BlockedUserId),
		)
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.USER_NOT_BLOCKED, "user not blocked"), "userId=%d, blockedUserId=%d", in.UserId, in.BlockedUserId)
	}

	// 使用事务删除拉黑记录
	err = l.svcCtx.ImUserBlacklistModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		return l.svcCtx.ImUserBlacklistModel.RemoveBlocked(ctx, session, in.UserId, in.BlockedUserId)
	})
	if err != nil {
		l.Logger.Errorw("Remove blocked record failed",
			logx.Field("userId", in.UserId),
			logx.Field("blockedUserId", in.BlockedUserId),
			logx.Field("error", err.Error()),
		)
		return nil, errors.Wrapf(err, "remove blocked record failed")
	}

	l.Logger.Infow("User unblocked successfully",
		logx.Field("userId", in.UserId),
		logx.Field("blockedUserId", in.BlockedUserId),
	)

	return &friend.UnblockUserResp{
		Success: true,
	}, nil
}
