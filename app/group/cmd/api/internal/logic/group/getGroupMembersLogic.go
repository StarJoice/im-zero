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

type GetGroupMembersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取群成员列表
func NewGetGroupMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupMembersLogic {
	return &GetGroupMembersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGroupMembersLogic) GetGroupMembers(req *types.GetGroupMembersReq) (resp *types.GetGroupMembersResp, err error) {
	// 参数验证
	if req.GroupId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid group id"), "groupId=%d", req.GroupId)
	}

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

	// 调用RPC服务获取群成员列表
	rpcResp, err := l.svcCtx.GroupRpc.GetGroupMembers(l.ctx, &group.GetGroupMembersReq{
		GroupId: req.GroupId,
		Role:    req.Role,
		Page:    req.Page,
		Limit:   req.Limit,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "get group members from rpc failed")
	}

	// 转换返回结果
	members := make([]types.GroupMember, 0, len(rpcResp.Members))
	for _, member := range rpcResp.Members {
		members = append(members, types.GroupMember{
			UserId:      member.UserId,
			GroupId:     member.GroupId,
			Nickname:    member.Nickname,
			Avatar:      member.Avatar,
			Role:        member.Role,
			Status:      member.Status,
			MuteEndTime: member.MuteEndTime,
			JoinTime:    member.JoinTime,
		})
	}

	return &types.GetGroupMembersResp{
		Members: members,
		Total:   rpcResp.Total,
	}, nil
}
