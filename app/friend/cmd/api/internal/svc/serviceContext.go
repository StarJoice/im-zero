package svc

import (
	"im-zero/app/friend/cmd/api/internal/config"
	"im-zero/app/friend/cmd/rpc/friendClient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	FriendRpc friendClient.Friend
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		FriendRpc: friendClient.NewFriend(zrpc.MustNewClient(c.FriendRpc)),
	}
}
