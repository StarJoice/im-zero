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

type GetGroupInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取群组信息
func NewGetGroupInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupInfoLogic {
	return &GetGroupInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGroupInfoLogic) GetGroupInfo(req *types.GetGroupInfoReq) (resp *types.GetGroupInfoResp, err error) {
	// 参数验证
	if req.GroupId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid group id"), "groupId=%d", req.GroupId)
	}

	// 获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.UNAUTHORIZED, "user not login"), "userId is empty")
	}

	// 调用RPC服务获取群组信息
	rpcResp, err := l.svcCtx.GroupRpc.GetGroupInfo(l.ctx, &group.GetGroupInfoReq{
		GroupId: req.GroupId,
		UserId:  userId,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get group info from rpc failed")
	}

	// 转换返回结果
	return &types.GetGroupInfoResp{
		Group: types.GroupInfo{
			Id:                rpcResp.Group.Id,
			Name:              rpcResp.Group.Name,
			Avatar:            rpcResp.Group.Avatar,
			Description:       rpcResp.Group.Description,
			Notice:            rpcResp.Group.Notice,
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
		MyRole: rpcResp.MyRole,
	}, nil
}
