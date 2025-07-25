package svc

import (
	"im-zero/app/message/cmd/api/internal/config"
	"im-zero/app/message/cmd/rpc/messageClient"
	
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	
	// RPC客户端
	MessageRpc messageClient.Message
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		
		// 初始化RPC客户端
		MessageRpc: messageClient.NewMessage(zrpc.MustNewClient(c.MessageRpc)),
	}
}
