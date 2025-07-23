package svc

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
	"im-zero/app/usercenter/cmd/rpc/internal/config"
	"im-zero/app/usercenter/model"
	"im-zero/app/verifycode/cmd/rpc/verifycode"
)

type ServiceContext struct {
	Config      config.Config
	RedisClient *redis.Redis

	UserModel     model.UserModel
	UserAuthModel model.UserAuthModel
	VerifycodeRpc verifycode.Verifycode
}

func NewServiceContext(c config.Config) *ServiceContext {

	sqlConn := sqlx.NewMysql(c.DB.DataSource)

	return &ServiceContext{
		Config: c,
		RedisClient: redis.New(c.Redis.Host, func(r *redis.Redis) {
			r.Type = c.Redis.Type
			r.Pass = c.Redis.Pass
		}),

		UserAuthModel: model.NewUserAuthModel(sqlConn, c.Cache),
		UserModel:     model.NewUserModel(sqlConn, c.Cache),
		VerifycodeRpc: verifycode.NewVerifycode(zrpc.MustNewClient(c.VerifycodeRpc)),
	}
}
