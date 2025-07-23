package logic

import (
	"context"
	"github.com/jinzhu/copier"
	"im-zero/app/usercenter/cmd/rpc/usercenter"

	"im-zero/app/usercenter/cmd/rpc/internal/svc"
	"im-zero/app/usercenter/cmd/rpc/pb"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *pb.GetUserInfoReq) (*pb.GetUserInfoResp, error) {
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.Wrapf(ErrUserNoExistsError, "id:%d", in.Id)
	}
	var respUser usercenter.User
	_ = copier.Copy(&respUser, user)

	return &usercenter.GetUserInfoResp{
		User: &respUser,
	}, nil
}
