package group

import (
	"context"

	"im-zero/app/group/cmd/api/internal/svc"
	"im-zero/app/group/cmd/api/internal/types"
	"im-zero/app/group/cmd/rpc/group"
	"im-zero/pkg/ctxdata"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyGroupsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取我的群组列表
func NewGetMyGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyGroupsLogic {
	return &GetMyGroupsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyGroupsLogic) GetMyGroups(req *types.GetMyGroupsReq) (resp *types.GetMyGroupsResp, err error) {
	// 获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.UNAUTHORIZED, "user not login"), "userId is empty")
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	// 调用RPC服务获取用户群组列表
	rpcResp, err := l.svcCtx.GroupRpc.GetUserGroups(l.ctx, &group.GetUserGroupsReq{
		UserId: userId,
		Page:   req.Page,
		Limit:  req.Limit,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get user groups from rpc failed")
	}

	// 转换返回结果
	groups := make([]types.GroupInfo, 0, len(rpcResp.Groups))
	for _, g := range rpcResp.Groups {
		groups = append(groups, types.GroupInfo{
			Id:                g.Id,
			Name:              g.Name,
			Avatar:            g.Avatar,
			Description:       g.Description,
			Notice:            g.Notice,
			OwnerId:           g.OwnerId,
			MemberCount:       g.MemberCount,
			MaxMembers:        g.MaxMembers,
			Status:            g.Status,
			IsPrivate:         g.IsPrivate,
			JoinApproval:      g.JoinApproval,
			AllowInvite:       g.AllowInvite,
			AllowMemberModify: g.AllowMemberModify,
			CreateTime:        g.CreateTime,
			UpdateTime:        g.UpdateTime,
		})
	}

	return &types.GetMyGroupsResp{
		Groups: groups,
		Total:  rpcResp.Total,
	}, nil
}
