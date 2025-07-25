package user

import (
	"context"

	"im-zero/app/usercenter/cmd/api/internal/svc"
	"im-zero/app/usercenter/cmd/api/internal/types"
	"im-zero/app/usercenter/cmd/rpc/usercenter"
	"im-zero/app/usercenter/model"
	"im-zero/pkg/tool"
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
	// 参数验证
	if err := l.validateLoginParams(req); err != nil {
		return nil, err
	}

	loginResp, err := l.svcCtx.UsercenterRpc.Login(l.ctx, &usercenter.LoginReq{
		AuthKey:  req.Mobile,
		AuthType: model.UserAuthTypeSystem,
		Password: req.Password,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "登录失败: mobile:%s", req.Mobile)
	}

	// 记录登录成功日志
	l.Logger.Infof("User login successfully: mobile=%s", req.Mobile)

	return &types.LoginResp{
		AccessToken:  loginResp.AccessToken,
		AccessExpire: loginResp.AccessExpire,
		RefreshAfter: loginResp.RefreshAfter,
	}, nil
}

// validateLoginParams 验证登录参数
func (l *LoginLogic) validateLoginParams(req *types.LoginReq) error {
	// 手机号验证
	if req.Mobile == "" {
		return errors.Wrapf(xerrs.NewErrCode(xerrs.PARAM_ERROR), "手机号不能为空")
	}
	if !tool.ValidateMobile(req.Mobile) {
		return errors.Wrapf(xerrs.NewErrCode(xerrs.PARAM_ERROR), "手机号格式不正确")
	}

	// 密码验证
	if req.Password == "" {
		return errors.Wrapf(xerrs.NewErrCode(xerrs.PARAM_ERROR), "密码不能为空")
	}
	if len(req.Password) < 6 {
		return errors.Wrapf(xerrs.NewErrCode(xerrs.PARAM_ERROR), "密码长度不能少于6位")
	}
	if len(req.Password) > 20 {
		return errors.Wrapf(xerrs.NewErrCode(xerrs.PARAM_ERROR), "密码长度不能超过20位")
	}

	return nil
}
