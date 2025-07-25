package svc

import (
	"im-zero/app/message/cmd/rpc/internal/config"
	"im-zero/app/message/model"
	"im-zero/app/usercenter/cmd/rpc/usercenter"
	"im-zero/app/friend/cmd/rpc/friendClient"
	"im-zero/app/group/cmd/rpc/groupClient"
	
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	
	// 数据模型
	ImMessageModel      model.ImMessageModel
	ImConversationModel model.ImConversationModel
	
	// RPC客户端
	UsercenterRpc usercenter.Usercenter
	FriendRpc     friendClient.Friend
	GroupRpc      groupClient.Group
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化数据库连接
	conn := sqlx.NewMysql(c.DB.DataSource)
	
	return &ServiceContext{
		Config: c,
		
		// 初始化数据模型
		ImMessageModel:      model.NewImMessageModel(conn, c.Cache),
		ImConversationModel: model.NewImConversationModel(conn, c.Cache),
		
		// 初始化RPC客户端
		UsercenterRpc: usercenter.NewUsercenter(zrpc.MustNewClient(c.UsercenterRpc)),
		FriendRpc:     friendClient.NewFriend(zrpc.MustNewClient(c.FriendRpc)),
		GroupRpc:      groupClient.NewGroup(zrpc.MustNewClient(c.GroupRpc)),
	}
}
