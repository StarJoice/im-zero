package logic

import (
	"context"

	"im-zero/app/friend/cmd/rpc/friend"
	"im-zero/app/friend/cmd/rpc/internal/svc"
	"im-zero/app/usercenter/cmd/rpc/usercenter"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetFriendListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFriendListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendListLogic {
	return &GetFriendListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取好友列表
func (l *GetFriendListLogic) GetFriendList(in *friend.GetFriendListReq) (*friend.GetFriendListResp, error) {
	// 记录请求开始
	l.Logger.Infow("Start getting friend list",
		logx.Field("userId", in.UserId),
		logx.Field("groupId", in.GroupId),
	)

	// 参数验证
	if in.UserId <= 0 {
		l.Logger.Errorw("Invalid user id", logx.Field("userId", in.UserId))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", in.UserId)
	}

	// 查询好友列表
	friends, err := l.svcCtx.ImFriendModel.FindByUserId(l.ctx, in.UserId)
	if err != nil {
		l.Logger.Errorw("Find friends failed", 
			logx.Field("userId", in.UserId), 
			logx.Field("error", err.Error()))
		return nil, errors.Wrapf(err, "find friends failed")
	}

	// 转换结果
	var friendList []*friend.FriendInfo
	for _, friendModel := range friends {
		// 过滤分组（如果指定了groupId）
		if in.GroupId > 0 && friendModel.GroupId.Int64 != in.GroupId {
			continue
		}

		// 获取好友用户信息
		userInfo, err := l.svcCtx.UsercenterRpc.GetUserInfo(l.ctx, &usercenter.GetUserInfoReq{
			Id: friendModel.FriendId,
		})
		if err != nil {
			l.Logger.Warnw("Get friend user info failed",
				logx.Field("friendId", friendModel.FriendId),
				logx.Field("error", err.Error()),
			)
			continue // 跳过获取失败的用户
		}

		friendInfo := &friend.FriendInfo{
			Id:         friendModel.Id,
			UserId:     friendModel.UserId,
			FriendId:   friendModel.FriendId,
			Remark:     friendModel.Remark.String,
			GroupId:    friendModel.GroupId.Int64,
			Status:     int32(friendModel.Status),
			CreateTime: friendModel.CreateTime.Unix(),
			UserInfo: &friend.UserInfo{
				Id:       userInfo.User.Id,
				Mobile:   userInfo.User.Mobile,
				Nickname: userInfo.User.Nickname,
				Avatar:   userInfo.User.Avatar,
				Sign:     userInfo.User.Sign,
			},
		}
		friendList = append(friendList, friendInfo)
	}

	l.Logger.Infow("Get friend list successfully",
		logx.Field("userId", in.UserId),
		logx.Field("totalCount", len(friends)),
		logx.Field("filteredCount", len(friendList)),
	)

	return &friend.GetFriendListResp{
		Friends: friendList,
		Total:   int32(len(friendList)),
	}, nil
}
