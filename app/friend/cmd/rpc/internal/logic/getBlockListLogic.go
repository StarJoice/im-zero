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

type GetBlockListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBlockListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlockListLogic {
	return &GetBlockListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取黑名单
func (l *GetBlockListLogic) GetBlockList(in *friend.GetBlockListReq) (*friend.GetBlockListResp, error) {
	// 记录获取黑名单请求开始
	l.Logger.Infow("Start getting block list",
		logx.Field("userId", in.UserId),
		logx.Field("page", in.Page),
		logx.Field("limit", in.Limit),
	)

	// 参数验证
	if in.UserId <= 0 {
		l.Logger.Errorw("Invalid user id", logx.Field("userId", in.UserId))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", in.UserId)
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.Limit <= 0 || in.Limit > 50 {
		in.Limit = 20
	}

	// 获取黑名单列表
	blockList, err := l.svcCtx.ImUserBlacklistModel.FindBlacklistByUserId(l.ctx, in.UserId)
	if err != nil {
		l.Logger.Errorw("Find block list failed",
			logx.Field("userId", in.UserId),
			logx.Field("error", err.Error()),
		)
		return nil, errors.Wrapf(err, "find block list failed")
	}

	// 分页处理
	total := len(blockList)
	start := (int(in.Page) - 1) * int(in.Limit)
	end := start + int(in.Limit)
	if start >= total {
		blockList = []*model.ImUserBlacklist{}
	} else {
		if end > total {
			end = total
		}
		blockList = blockList[start:end]
	}

	// 转换结果
	var blockedUsers []*friend.UserInfo
	for _, blocked := range blockList {
		// 获取被拉黑用户信息
		userInfo, err := l.svcCtx.UsercenterRpc.GetUserInfo(l.ctx, &usercenter.GetUserInfoReq{
			Id: blocked.BlockedUserId,
		})
		if err != nil {
			l.Logger.Errorw("Get blocked user info failed",
				logx.Field("blockedUserId", blocked.BlockedUserId),
				logx.Field("error", err.Error()),
			)
			continue // 跳过获取失败的用户
		}

		blockedUserInfo := &friend.UserInfo{
			Id:       userInfo.User.Id,
			Mobile:   userInfo.User.Mobile,
			Nickname: userInfo.User.Nickname,
			Avatar:   userInfo.User.Avatar,
			Sign:     userInfo.User.Sign,
			Status:   0, // 状态需要从其他服务获取，暂时设为0
		}
		blockedUsers = append(blockedUsers, blockedUserInfo)
	}

	l.Logger.Infow("Get block list successfully",
		logx.Field("userId", in.UserId),
		logx.Field("total", total),
		logx.Field("returnedCount", len(blockedUsers)),
	)

	return &friend.GetBlockListResp{
		BlockedUsers: blockedUsers,
		Total:        int32(total),
	}, nil
}
