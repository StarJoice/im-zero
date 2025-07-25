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

type UpdateGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新群组信息
func NewUpdateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGroupLogic {
	return &UpdateGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateGroupLogic) UpdateGroup(req *types.UpdateGroupReq) (resp *types.UpdateGroupResp, err error) {
	// 参数验证
	if req.GroupId <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid group id"), "groupId=%d", req.GroupId)
	}

	// 获取当前用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId == 0 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.UNAUTHORIZED, "user not login"), "userId is empty")
	}

	// 调用RPC服务更新群组信息
	rpcResp, err := l.svcCtx.GroupRpc.UpdateGroup(l.ctx, &group.UpdateGroupReq{
		GroupId:           req.GroupId,
		OperatorId:        userId,
		Name:              req.Name,
		Avatar:            req.Avatar,
		Description:       req.Description,
		Notice:            req.Notice,
		JoinApproval:      req.JoinApproval,
		AllowInvite:       req.AllowInvite,
		AllowMemberModify: req.AllowMemberModify,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "update group from rpc failed")
	}

	return &types.UpdateGroupResp{
		Success: rpcResp.Success,
	}, nil
}
