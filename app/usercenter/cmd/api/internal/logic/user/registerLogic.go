package user

import (
	"context"
	"im-zero/app/usercenter/cmd/rpc/usercenter"
	"im-zero/app/usercenter/model"
	"im-zero/pkg/tool"
	"im-zero/pkg/xerrs"

	"im-zero/app/usercenter/cmd/api/internal/svc"
	"im-zero/app/usercenter/cmd/api/internal/types"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// register
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	// 参数验证
	if err := l.validateRegisterParams(req); err != nil {
		return nil, err
	}

	registerResp, err := l.svcCtx.UsercenterRpc.Register(l.ctx, &usercenter.RegisterReq{
		Mobile:   req.Mobile,
		Password: req.Password,
		Code:     req.Code,
		CodeKey:  req.CodeKey,
		AuthKey:  req.Mobile,
		AuthType: model.UserAuthTypeSystem,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "注册失败: %+v", req)
	}

	// 检查复制是否成功
	resp = &types.RegisterResp{}
	if err = copier.Copy(resp, registerResp); err != nil {
		return nil, errors.Wrapf(xerrs.NewErrCode(xerrs.SERVER_COMMON_ERROR), "copy response failed: %v", err)
	}

	// 记录注册成功日志
	l.Logger.Infof("User registered successfully: mobile=%s", req.Mobile)

	return resp, nil
}

// validateRegisterParams 验证注册参数
func (l *RegisterLogic) validateRegisterParams(req *types.RegisterReq) error {
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

	// 验证码验证
	if req.Code == "" {
		return errors.Wrapf(xerrs.NewErrCode(xerrs.PARAM_ERROR), "验证码不能为空")
	}
	if req.CodeKey == "" {
		return errors.Wrapf(xerrs.NewErrCode(xerrs.PARAM_ERROR), "验证码Key不能为空")
	}
	if len(req.Code) != 6 {
		return errors.Wrapf(xerrs.NewErrCode(xerrs.PARAM_ERROR), "验证码格式不正确")
	}

	return nil
}
