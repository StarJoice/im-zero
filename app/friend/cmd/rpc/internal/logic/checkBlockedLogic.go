package logic

import (
	"context"

	"im-zero/app/friend/cmd/rpc/friend"
	"im-zero/app/friend/cmd/rpc/internal/svc"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type CheckBlockedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckBlockedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckBlockedLogic {
	return &CheckBlockedLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 检查是否被拉黑
func (l *CheckBlockedLogic) CheckBlocked(in *friend.CheckBlockedReq) (*friend.CheckBlockedResp, error) {
	// 记录检查拉黑状态请求开始
	l.Logger.Infow("Start checking blocked status",
		logx.Field("userId", in.UserId),
		logx.Field("targetUserId", in.TargetUserId),
	)

	// 参数验证
	if in.UserId <= 0 {
		l.Logger.Errorw("Invalid user id", logx.Field("userId", in.UserId))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", in.UserId)
	}
	if in.TargetUserId <= 0 {
		l.Logger.Errorw("Invalid target user id", logx.Field("targetUserId", in.TargetUserId))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid target user id"), "targetUserId=%d", in.TargetUserId)
	}

	// 检查双向拉黑状态
	isBlocked, hasBlocked, err := l.svcCtx.ImUserBlacklistModel.CheckMutualBlocked(l.ctx, in.UserId, in.TargetUserId)
	if err != nil {
		l.Logger.Errorw("Check mutual blocked status failed",
			logx.Field("userId", in.UserId),
			logx.Field("targetUserId", in.TargetUserId),
			logx.Field("error", err.Error()),
		)
		return nil, errors.Wrapf(err, "check mutual blocked status failed")
	}

	l.Logger.Infow("Check blocked status completed",
		logx.Field("userId", in.UserId),
		logx.Field("targetUserId", in.TargetUserId),
		logx.Field("isBlocked", isBlocked),
		logx.Field("hasBlocked", hasBlocked),
	)

	return &friend.CheckBlockedResp{
		IsBlocked:  isBlocked,  // userId 是否被 targetUserId 拉黑
		HasBlocked: hasBlocked, // userId 是否拉黑了 targetUserId
	}, nil
}
