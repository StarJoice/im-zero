package logic

import (
	"context"
	"im-zero/app/usercenter/cmd/rpc/internal/svc"
	"im-zero/app/usercenter/cmd/rpc/pb"
	"im-zero/app/usercenter/cmd/rpc/usercenter"
	"im-zero/app/usercenter/model"
	"im-zero/app/verifycode/cmd/rpc/verifycode"
	"im-zero/pkg/tool"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var ErrUserAlreadyRegisterError = xerrs.NewErrMsg("user has been registered")
var ErrGenerateTokenError = xerrs.NewErrMsg("generate token failed")

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *pb.RegisterReq) (*pb.RegisterResp, error) {
	if in.Mobile == "" || in.Password == "" || in.Code == "" || in.CodeKey == "" {
		return nil, errors.Wrapf(xerrs.NewErrCode(xerrs.PARAM_ERROR), "Parameter cannot be empty")
	}

	// 验证手机号是否已经注册
	user, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, in.Mobile)
	if err != nil && err != model.ErrNotFound {
		return nil, errors.Wrapf(xerrs.NewErrCode(xerrs.DB_ERROR), "Failed to check mobile existence: %v, mobile: %s", err, in.Mobile)
	}
	if user != nil {
		return nil, errors.Wrapf(ErrUserAlreadyRegisterError, "User already exists for mobile: %s", in.Mobile)
	}

	// 验证已有的验证码
	resp, err := l.svcCtx.VerifycodeRpc.VerifySmsCode(l.ctx, &verifycode.VerifySmsCodeReq{
		Mobile:  in.Mobile,
		Code:    in.Code,
		CodeKey: in.CodeKey,
		Scene:   1,
	})
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, errors.Wrap(xerrs.NewErrCode(xerrs.VERIFY_CODE_ERROR), "Invalid verification code")
	}

	var userId int64
	if err := l.svcCtx.UserModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 创建用户记录
		user := new(model.User)
		user.Mobile = in.Mobile
		user.Nickname = tool.Krand(8, tool.KC_RAND_KIND_ALL)
		user.Password, _ = tool.HashPassword(in.Password)

		insertResult, err := l.svcCtx.UserModel.Insert(ctx, session, user)
		if err != nil {
			return errors.Wrapf(xerrs.NewErrCode(xerrs.DB_ERROR), "Failed to insert user: %v, user: %+v", err, user)
		}

		lastId, err := insertResult.LastInsertId()
		if err != nil {
			return errors.Wrapf(xerrs.NewErrCode(xerrs.DB_ERROR), "Failed to get last insert ID: %v, user: %+v", err, user)
		}

		userId = lastId

		// 创建用户认证记录
		userAuth := new(model.UserAuth)
		userAuth.UserId = lastId
		userAuth.AuthKey = in.AuthKey
		userAuth.AuthType = in.AuthType

		if _, err := l.svcCtx.UserAuthModel.Insert(ctx, session, userAuth); err != nil {
			return errors.Wrapf(xerrs.NewErrCode(xerrs.DB_ERROR), "Failed to insert user auth: %v, userAuth: %v", err, userAuth)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// 生成访问令牌
	generateTokenLogic := NewGenerateTokenLogic(l.ctx, l.svcCtx)
	tokenResp, err := generateTokenLogic.GenerateToken(&usercenter.GenerateTokenReq{
		UserId: userId,
	})
	if err != nil {
		return nil, errors.Wrapf(ErrGenerateTokenError, "Failed to generate token for user ID: %d", userId)
	}

	return &usercenter.RegisterResp{
		AccessToken:  tokenResp.AccessToken,
		AccessExpire: tokenResp.AccessExpire,
		RefreshAfter: tokenResp.RefreshAfter,
	}, nil
}
