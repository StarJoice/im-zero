package user

import (
	"context"

	"im-zero/app/usercenter/cmd/api/internal/svc"
	"im-zero/app/usercenter/cmd/api/internal/types"
	"im-zero/app/usercenter/cmd/rpc/usercenter"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type DetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// get user info
func NewDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DetailLogic {
	return &DetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DetailLogic) Detail(req *types.UserInfoReq) (resp *types.UserInfoResp, err error) {
	if req.Id <= 0 {
		return nil, errors.Wrapf(xerrs.NewErrCode(xerrs.PARAM_ERROR), "用户ID不能为空")
	}

	userInfoResp, err := l.svcCtx.UsercenterRpc.GetUserInfo(l.ctx, &usercenter.GetUserInfoReq{
		Id: req.Id,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "获取用户信息失败: id:%d", req.Id)
	}

	return &types.UserInfoResp{
		User: types.User{
			Id:       userInfoResp.User.Id,
			Mobile:   userInfoResp.User.Mobile,
			Nickname: userInfoResp.User.Nickname,
			Avatar:   userInfoResp.User.Avatar,
			Sign:     userInfoResp.User.Sign,
			Info:     userInfoResp.User.Info,
		},
	}, nil
}
