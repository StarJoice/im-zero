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
	friends, err := l.svcCtx.ImFriendModel.FindFriendsByUserId(l.ctx, in.UserId)
	if err != nil {
		l.Logger.Errorw("Find friends failed",
			logx.Field("userId", in.UserId),
			logx.Field("error", err.Error()))
		return nil, errors.Wrapf(err, "find friends failed")
	}

	// 转换结果
	var friendList []*friend.FriendInfo
	for _, friendModel := range friends {
		// 获取好友用户信息
		userInfo, err := l.svcCtx.UsercenterRpc.GetUserInfo(l.ctx, &usercenter.GetUserInfoReq{
			Id: friendModel.FriendId,
		})
		if err != nil {
			l.Logger.Errorw("Get friend user info failed",
				logx.Field("friendId", friendModel.FriendId),
				logx.Field("error", err.Error()),
			)
			continue // 跳过获取失败的用户
		}

		friendInfo := &friend.FriendInfo{
			UserInfo: &friend.UserInfo{
				Id:       userInfo.User.Id,
				Mobile:   userInfo.User.Mobile,
				Nickname: userInfo.User.Nickname,
				Avatar:   userInfo.User.Avatar,
				Sign:     userInfo.User.Sign,
				Status:   0, // 在线状态需要从其他服务获取，暂时设为0
			},
			Remark:     friendModel.Remark.String,
			GroupId:    0, // 当前数据模型中没有GroupId字段，设为0
			CreateTime: friendModel.CreateTime.Unix(),
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
