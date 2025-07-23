package user

import (
	"context"
	"im-zero/app/usercenter/cmd/rpc/usercenter"
	"im-zero/app/usercenter/model"

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

	_ = copier.Copy(&resp, registerResp)
	return resp, nil
}
