package verifycode

import (
	"context"
	"im-zero/app/verifycode/cmd/rpc/verifycode"

	"im-zero/app/verifycode/cmd/api/internal/svc"
	"im-zero/app/verifycode/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendSmsCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// send verifycode
func NewSendSmsCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendSmsCodeLogic {
	return &SendSmsCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendSmsCodeLogic) SendSmsCode(req *types.SendSmsCodeReq) (resp *types.SendSmsCodeResp, err error) {
	rpcResp, err := l.svcCtx.VerifycodeRpc.SendSmsCode(l.ctx, &verifycode.SendSmsCodeReq{
		Mobile: req.Mobile,
		Scene:  req.Scene,
	})
	if err != nil {
		return nil, err
	}

	return &types.SendSmsCodeResp{
		CodeKey: rpcResp.CodeKey,
	}, nil
}
