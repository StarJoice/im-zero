package friend

import (
	"context"

	"im-zero/app/friend/cmd/api/internal/svc"
	"im-zero/app/friend/cmd/api/internal/types"
	"im-zero/app/friend/cmd/rpc/friend"
	"im-zero/pkg/ctxdata"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetFriendListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取好友列表
func NewGetFriendListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendListLogic {
	return &GetFriendListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFriendListLogic) GetFriendList(req *types.GetFriendListReq) (resp *types.GetFriendListResp, err error) {
	// 获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.UNAUTHORIZED, "user not login"), "userId is empty")
	}

	// 调用RPC服务获取好友列表
	rpcResp, err := l.svcCtx.FriendRpc.GetFriendList(l.ctx, &friend.GetFriendListReq{
		UserId:  userId,
		GroupId: req.GroupId,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get friend list rpc failed")
	}

	// 转换返回结果
	var friends []types.FriendInfo
	for _, friendInfo := range rpcResp.Friends {
		friends = append(friends, types.FriendInfo{
			Id:         friendInfo.Id,
			UserId:     friendInfo.UserId,
			FriendId:   friendInfo.FriendId,
			Remark:     friendInfo.Remark,
			GroupId:    friendInfo.GroupId,
			Status:     friendInfo.Status,
			CreateTime: friendInfo.CreateTime,
			// 用户信息
			UserInfo: types.UserInfo{
				Id:       friendInfo.UserInfo.Id,
				Mobile:   friendInfo.UserInfo.Mobile,
				Nickname: friendInfo.UserInfo.Nickname,
				Avatar:   friendInfo.UserInfo.Avatar,
				Sign:     friendInfo.UserInfo.Sign,
			},
		})
	}

	// 记录操作日志
	l.Logger.Infof("Get friend list successfully: userId=%d, count=%d", userId, len(friends))

	return &types.GetFriendListResp{
		Friends: friends,
		Total:   rpcResp.Total,
	}, nil
}
