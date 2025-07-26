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

type GetBlockListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取黑名单
func NewGetBlockListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlockListLogic {
	return &GetBlockListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBlockListLogic) GetBlockList(req *types.GetBlockListReq) (resp *types.GetBlockListResp, err error) {
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 50 {
		req.Limit = 50
	}

	// 获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.UNAUTHORIZED, "user not login"), "userId is empty")
	}

	// 调用RPC服务获取黑名单
	rpcResp, err := l.svcCtx.FriendRpc.GetBlockList(l.ctx, &friend.GetBlockListReq{
		UserId: userId,
		Page:   req.Page,
		Limit:  req.Limit,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get block list rpc failed")
	}

	// 转换返回结果
	var blockUsers []types.UserInfo
	for _, userInfo := range rpcResp.BlockedUsers {
		blockUsers = append(blockUsers, types.UserInfo{
			Id:       userInfo.Id,
			Mobile:   userInfo.Mobile,
			Nickname: userInfo.Nickname,
			Avatar:   userInfo.Avatar,
			Sign:     userInfo.Sign,
			Status:   userInfo.Status,
		})
	}

	// 记录操作日志
	l.Logger.Infof("Get block list successfully: userId=%d, count=%d", userId, len(blockUsers))

	return &types.GetBlockListResp{
		BlockUsers: blockUsers,
		Total:      rpcResp.Total,
	}, nil
}
