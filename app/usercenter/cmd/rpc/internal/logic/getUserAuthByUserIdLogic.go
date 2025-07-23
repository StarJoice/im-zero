package logic

import (
	"context"
	"github.com/jinzhu/copier"
	"im-zero/app/usercenter/cmd/rpc/usercenter"
	"im-zero/app/usercenter/model"
	"im-zero/pkg/xerrs"

	"im-zero/app/usercenter/cmd/rpc/internal/svc"
	"im-zero/app/usercenter/cmd/rpc/pb"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserAuthByUserIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserAuthByUserIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserAuthByUserIdLogic {
	return &GetUserAuthByUserIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserAuthByUserIdLogic) GetUserAuthByUserId(in *pb.GetUserAuthByUserIdReq) (*pb.GetUserAuthyUserIdResp, error) {
	// todo: add your logic here and delete this line
	userAuth, err := l.svcCtx.UserAuthModel.FindOneByUserIdAuthType(l.ctx, in.UserId, in.AuthType)
	if err != nil && err != model.ErrNotFound {
		return nil, errors.Wrapf(xerrs.NewErrMsg("get user auth  fail"), "err : %v , in : %+v", err, in)
	}

	var respUserAuth usercenter.UserAuth
	_ = copier.Copy(&respUserAuth, userAuth)

	return &pb.GetUserAuthyUserIdResp{
		UserAuth: &respUserAuth,
	}, nil

}
