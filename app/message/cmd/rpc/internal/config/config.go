package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DB struct {
		DataSource string
	}
	Cache cache.CacheConf
	
	// 依赖的其他RPC服务
	UsercenterRpc zrpc.RpcClientConf
	FriendRpc     zrpc.RpcClientConf
	GroupRpc      zrpc.RpcClientConf
}
