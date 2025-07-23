package svc

import (
	"github.com/zeromicro/go-zero/zrpc"
	"im-zero/app/verifycode/cmd/api/internal/config"
	"im-zero/app/verifycode/cmd/rpc/verifycode"
)

type ServiceContext struct {
	Config config.Config
	// / 注入rpc服务
	VerifycodeRpc verifycode.Verifycode
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:        c,
		VerifycodeRpc: verifycode.NewVerifycode(zrpc.MustNewClient(c.VerifycodeRpc)),
	}
}
