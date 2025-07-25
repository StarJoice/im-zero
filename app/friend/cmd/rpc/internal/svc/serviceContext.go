package svc

import (
	"im-zero/app/friend/cmd/rpc/internal/config"
	"im-zero/app/friend/model"
	"im-zero/app/usercenter/cmd/rpc/usercenter"
	
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	
	// 数据模型
	ImFriendModel        model.ImFriendModel
	ImFriendRequestModel model.ImFriendRequestModel
	ImUserBlacklistModel model.ImUserBlacklistModel
	
	// RPC客户端
	UsercenterRpc usercenter.Usercenter
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化数据库连接
	conn := sqlx.NewMysql(c.DB.DataSource)
	
	return &ServiceContext{
		Config: c,
		
		// 初始化数据模型
		ImFriendModel:        model.NewImFriendModel(conn, c.Cache),
		ImFriendRequestModel: model.NewImFriendRequestModel(conn, c.Cache),
		ImUserBlacklistModel: model.NewImUserBlacklistModel(conn, c.Cache),
		
		// 初始化RPC客户端
		UsercenterRpc: usercenter.NewUsercenter(zrpc.MustNewClient(c.UsercenterRpc)),
	}
}
