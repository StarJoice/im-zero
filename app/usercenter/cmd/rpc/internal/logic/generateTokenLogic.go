package logic

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"time"

	"im-zero/app/usercenter/cmd/rpc/internal/svc"
	"im-zero/app/usercenter/cmd/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateTokenLogic {
	return &GenerateTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}
func (l *GenerateTokenLogic) GenerateToken(in *pb.GenerateTokenReq) (*pb.GenerateTokenResp, error) {
	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.JwtAuth.AccessExpire // 单位：秒

	accessToken, err := l.getJwtToken(
		l.svcCtx.Config.JwtAuth.AccessSecret,
		now,
		accessExpire,
		in.UserId,
	)

	if err != nil {
		logx.Errorf("生成Token失败 userID:%d error:%v", in.UserId, err)
		return nil, errors.Wrapf(ErrGenerateTokenError, "系统错误")
	}

	return &pb.GenerateTokenResp{
		AccessToken:  accessToken,
		AccessExpire: now + accessExpire,
		RefreshAfter: now + accessExpire/2,
	}, nil
}

type JwtClaims struct {
	UserID int64 `json:"userId"`
	jwt.RegisteredClaims
}

func (l *GenerateTokenLogic) getJwtToken(secretKey string, iat, seconds, userId int64) (string, error) {
	// 将秒级时间戳转换为time.Time
	iatTime := time.Unix(iat, 0)
	expTime := time.Unix(iat+seconds, 0)

	claims := JwtClaims{
		UserID: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(iatTime), // v5使用NumericDate
			ExpiresAt: jwt.NewNumericDate(expTime), // v5使用NumericDate
			// 可选添加其他标准声明
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}
