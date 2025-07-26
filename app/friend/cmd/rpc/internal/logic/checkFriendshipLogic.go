package logic

import (
	"context"

	"im-zero/app/friend/cmd/rpc/friend"
	"im-zero/app/friend/cmd/rpc/internal/svc"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type CheckFriendshipLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckFriendshipLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckFriendshipLogic {
	return &CheckFriendshipLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 检查好友关系
func (l *CheckFriendshipLogic) CheckFriendship(in *friend.CheckFriendshipReq) (*friend.CheckFriendshipResp, error) {
	// 记录检查好友关系请求开始
	l.Logger.Infow("Start checking friendship",
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

	// 自己和自己不能是好友关系
	if in.UserId == in.TargetUserId {
		l.Logger.Infow("User checking friendship with self",
			logx.Field("userId", in.UserId),
		)
		return &friend.CheckFriendshipResp{
			IsFriend: false,
		}, nil
	}

	// 检查好友关系
	isFriend, err := l.svcCtx.ImFriendModel.CheckFriendship(l.ctx, in.UserId, in.TargetUserId)
	if err != nil {
		l.Logger.Errorw("Check friendship failed",
			logx.Field("userId", in.UserId),
			logx.Field("targetUserId", in.TargetUserId),
			logx.Field("error", err.Error()),
		)
		return nil, errors.Wrapf(err, "check friendship failed")
	}

	l.Logger.Infow("Check friendship completed",
		logx.Field("userId", in.UserId),
		logx.Field("targetUserId", in.TargetUserId),
		logx.Field("isFriend", isFriend),
	)

	return &friend.CheckFriendshipResp{
		IsFriend: isFriend,
	}, nil
}
