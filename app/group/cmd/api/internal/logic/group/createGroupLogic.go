package group

import (
	"context"

	"im-zero/app/group/cmd/api/internal/svc"
	"im-zero/app/group/cmd/api/internal/types"
	"im-zero/app/group/cmd/rpc/group"
	"im-zero/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建群组
func NewCreateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGroupLogic {
	return &CreateGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateGroupLogic) CreateGroup(req *types.CreateGroupReq) (resp *types.CreateGroupResp, err error) {
	// 从上下文获取用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 调用RPC服务创建群组
	rpcResp, err := l.svcCtx.GroupRpc.CreateGroup(l.ctx, &group.CreateGroupReq{
		OwnerId:      userId,
		Name:         req.Name,
		Avatar:       req.Avatar,
		Description:  req.Description,
		MemberIds:    req.MemberIds,
		IsPrivate:    req.IsPrivate,
		JoinApproval: req.JoinApproval,
	})
	if err != nil {
		return nil, err
	}

	// 转换响应
	return &types.CreateGroupResp{
		Group: types.GroupInfo{
			Id:                rpcResp.Group.Id,
			Name:              rpcResp.Group.Name,
			Avatar:            rpcResp.Group.Avatar,
			Description:       rpcResp.Group.Description,
			OwnerId:           rpcResp.Group.OwnerId,
			MemberCount:       rpcResp.Group.MemberCount,
			MaxMembers:        rpcResp.Group.MaxMembers,
			Status:            rpcResp.Group.Status,
			IsPrivate:         rpcResp.Group.IsPrivate,
			JoinApproval:      rpcResp.Group.JoinApproval,
			AllowInvite:       rpcResp.Group.AllowInvite,
			AllowMemberModify: rpcResp.Group.AllowMemberModify,
			CreateTime:        rpcResp.Group.CreateTime,
			UpdateTime:        rpcResp.Group.UpdateTime,
		},
	}, nil
}
