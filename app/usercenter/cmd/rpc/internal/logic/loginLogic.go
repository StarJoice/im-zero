package logic

import (
	"context"
	"github.com/pkg/errors"
	"im-zero/app/usercenter/cmd/rpc/usercenter"
	"im-zero/app/usercenter/model"
	"im-zero/pkg/tool"
	"im-zero/pkg/xerrs"

	"im-zero/app/usercenter/cmd/rpc/internal/svc"
	"im-zero/app/usercenter/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

var ErrUserNoExistsError = xerrs.NewErrMsg("用户不存在")
var ErrUsernamePwdError = xerrs.NewErrMsg("账号或密码不正确")

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *pb.LoginReq) (*pb.LoginResp, error) {
	// 1.根据用户选择登陆方式，登录 todo 后续添加其他登陆方式
	var userId int64
	var err error
	switch in.AuthType {
	case model.UserAuthTypeSystem:
		userId, err = l.loginByMobile(in.AuthKey, in.Password)
	default:
		return nil, xerrs.NewErrCode(xerrs.SERVER_COMMON_ERROR)
	}
	if err != nil {
		return nil, err
	}
	//2. 生成token
	generateTokenLogic := NewGenerateTokenLogic(l.ctx, l.svcCtx)
	tokenResp, err := generateTokenLogic.GenerateToken(&usercenter.GenerateTokenReq{
		UserId: userId,
	})
	if err != nil {
		return nil, errors.Wrapf(ErrGenerateTokenError, "GenerateToken userId : %d", userId)
	}

	return &usercenter.LoginResp{
		AccessToken:  tokenResp.AccessToken,
		AccessExpire: tokenResp.AccessExpire,
		RefreshAfter: tokenResp.RefreshAfter,
	}, nil

}

func (l *LoginLogic) loginByMobile(mobile, password string) (int64, error) {
	user, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, mobile)
	if err != nil && err != model.ErrNotFound {
		return 0, errors.Wrapf(xerrs.NewErrCode(xerrs.DB_ERROR), "根据手机号查询用户失败，mobile:%s,err:%v", mobile, err)
	}
	if user == nil {
		return 0, errors.Wrapf(ErrUserNoExistsError, "mobile:%s", mobile)
	}

	if !tool.CheckPassword(password, user.Password) {
		return 0, errors.Wrap(ErrUsernamePwdError, "密码匹配出错")
	}
	return user.Id, nil
}

func (l *LoginLogic) loginByWx() error {
	return nil
}

func (l *LoginLogic) loginByVerifyCode(mobile, code string) (int64, error) {
	return 0, nil
}
