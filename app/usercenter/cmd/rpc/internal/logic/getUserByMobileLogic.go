package logic

import (
	"context"

	"im-zero/app/usercenter/cmd/rpc/internal/svc"
	"im-zero/app/usercenter/cmd/rpc/pb"
	"im-zero/app/usercenter/model"
	"im-zero/pkg/tool"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserByMobileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserByMobileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserByMobileLogic {
	return &GetUserByMobileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据手机号获取用户信息
func (l *GetUserByMobileLogic) GetUserByMobile(in *pb.GetUserByMobileReq) (*pb.GetUserByMobileResp, error) {
	// 参数验证
	if !tool.ValidateMobile(in.Mobile) {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.INVALID_MOBILE, "invalid mobile format"), "mobile=%s", in.Mobile)
	}

	// 查询用户
	user, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, in.Mobile)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.USER_NOT_FOUND, "user not found"), "mobile=%s", in.Mobile)
		}
		return nil, errors.Wrapf(err, "find user by mobile failed")
	}

	return &pb.GetUserByMobileResp{
		User: &pb.User{
			Id:       user.Id,
			Mobile:   user.Mobile,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Sign:     user.Sign,
		},
	}, nil
}
