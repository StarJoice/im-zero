package svc

import (
	"fmt"
	"github.com/zeromicro/go-zero/zrpc"
	"im-zero/app/usercenter/cmd/api/internal/config"
	"im-zero/app/usercenter/cmd/rpc/usercenter"
)

type ServiceContext struct {
	Config        config.Config
	UsercenterRpc usercenter.Usercenter
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 添加日志输出以调试
	fmt.Printf("RPC Config: %+v\n", c.UsercenterRpc)
	return &ServiceContext{
		Config:        c,
		UsercenterRpc: usercenter.NewUsercenter(zrpc.MustNewClient(c.UsercenterRpc)),
	}
}
