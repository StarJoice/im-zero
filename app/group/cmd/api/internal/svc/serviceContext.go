package svc

import (
	"im-zero/app/group/cmd/api/internal/config"
	"im-zero/app/group/cmd/rpc/groupClient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	// RPC客户端
	GroupRpc groupClient.Group
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 创建非阻塞的RPC客户端
	groupRpcClient, _ := zrpc.NewClient(c.GroupRpc)

	return &ServiceContext{
		Config: c,

		// 初始化RPC客户端（非阻塞方式）
		GroupRpc: groupClient.NewGroup(groupRpcClient),
	}
}
