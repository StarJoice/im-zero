package user

import (
	"context"

	"im-zero/app/usercenter/cmd/api/internal/svc"
	"im-zero/app/usercenter/cmd/api/internal/types"
	"im-zero/app/usercenter/cmd/rpc/usercenter"
	"im-zero/app/usercenter/model"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// login
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	if req.Mobile == "" || req.Password == "" {
		return nil, errors.Wrapf(xerrs.NewErrCode(xerrs.PARAM_ERROR), "手机号和密码不能为空")
	}

	loginResp, err := l.svcCtx.UsercenterRpc.Login(l.ctx, &usercenter.LoginReq{
		AuthKey:  req.Mobile,
		AuthType: model.UserAuthTypeSystem,
		Password: req.Password,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "登录失败: mobile:%s", req.Mobile)
	}

	return &types.LoginResp{
		AccessToken:  loginResp.AccessToken,
		AccessExpire: loginResp.AccessExpire,
		RefreshAfter: loginResp.RefreshAfter,
	}, nil
}
