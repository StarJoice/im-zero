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

type SearchUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 搜索用户
func NewSearchUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchUsersLogic {
	return &SearchUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchUsersLogic) SearchUsers(req *types.SearchUsersReq) (resp *types.SearchUsersResp, err error) {
	// 参数验证
	if len(req.Keyword) == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "keyword is required"), "keyword is empty")
	}
	if len(req.Keyword) < 2 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "keyword too short"), "keyword length=%d", len(req.Keyword))
	}
	if len(req.Keyword) > 50 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "keyword too long"), "keyword length=%d", len(req.Keyword))
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 20 {
		req.Limit = 20 // API定义中的最大限制
	}

	// 获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.UNAUTHORIZED, "user not login"), "userId is empty")
	}

	// 调用RPC服务搜索用户
	rpcResp, err := l.svcCtx.FriendRpc.SearchUsers(l.ctx, &friend.SearchUsersReq{
		Keyword: req.Keyword,
		Page:    req.Page,
		Limit:   req.Limit,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "search users rpc failed")
	}

	// 转换返回结果
	var users []types.UserInfo
	for _, userInfo := range rpcResp.Users {
		users = append(users, types.UserInfo{
			Id:       userInfo.Id,
			Mobile:   userInfo.Mobile,
			Nickname: userInfo.Nickname,
			Avatar:   userInfo.Avatar,
			Sign:     userInfo.Sign,
		})
	}

	// 记录操作日志
	l.Logger.Infof("Search users successfully: userId=%d, keyword=%s, count=%d", userId, req.Keyword, len(users))

	return &types.SearchUsersResp{
		Users: users,
		Total: rpcResp.Total,
	}, nil
}
