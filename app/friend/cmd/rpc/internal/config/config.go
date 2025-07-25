package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	
	// 数据库配置
	DB struct {
		DataSource string
	}
	
	// 缓存配置
	Cache cache.CacheConf
	
	// RPC配置
	UsercenterRpc zrpc.RpcClientConf
}
