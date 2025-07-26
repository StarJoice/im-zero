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
	// 创建非阻塞的RPC客户端
	messageRpcClient, _ := zrpc.NewClient(c.MessageRpc)

	return &ServiceContext{
		Config: c,

		// 初始化RPC客户端（非阻塞方式）
		MessageRpc: messageClient.NewMessage(messageRpcClient),
	}
}
