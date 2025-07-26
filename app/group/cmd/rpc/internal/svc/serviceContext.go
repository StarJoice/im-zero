package svc

import (
	"im-zero/app/group/cmd/rpc/internal/config"
	"im-zero/app/group/model"
	"im-zero/app/message/cmd/rpc/messageClient"
	"im-zero/app/usercenter/cmd/rpc/usercenter"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	// 数据模型
	ImGroupModel        model.ImGroupModel
	ImGroupMemberModel  model.ImGroupMemberModel
	ImGroupMessageModel model.ImGroupMessageModel
	ImGroupRequestModel model.ImGroupRequestModel

	// RPC客户端
	UsercenterRpc usercenter.Usercenter
	MessageRpc    messageClient.Message
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化数据库连接
	conn := sqlx.NewMysql(c.DB.DataSource)

	// 创建非阻塞的RPC客户端
	usercenterRpcClient, _ := zrpc.NewClient(c.UsercenterRpc)
	messageRpcClient, _ := zrpc.NewClient(c.MessageRpc)

	return &ServiceContext{
		Config: c,

		// 初始化数据模型
		ImGroupModel:        model.NewImGroupModel(conn, c.Cache),
		ImGroupMemberModel:  model.NewImGroupMemberModel(conn, c.Cache),
		ImGroupMessageModel: model.NewImGroupMessageModel(conn, c.Cache),
		ImGroupRequestModel: model.NewImGroupRequestModel(conn, c.Cache),

		// 初始化RPC客户端（非阻塞方式）
		UsercenterRpc: usercenter.NewUsercenter(usercenterRpcClient),
		MessageRpc:    messageClient.NewMessage(messageRpcClient),
	}
}
