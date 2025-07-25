package ws

import (
	"context"
	"net/http"
	"strconv"

	"im-zero/app/message/cmd/api/internal/svc"
	"im-zero/app/message/cmd/api/internal/types"
	"im-zero/pkg/ctxdata"
	"im-zero/pkg/wsmanager"

	"github.com/zeromicro/go-zero/core/logx"
)

type WsConnectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWsConnectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WsConnectLogic {
	return &WsConnectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WsConnectLogic) WsConnect(req *types.WsConnectReq, w http.ResponseWriter, r *http.Request) error {
	// 从JWT中获取用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)
	if userId <= 0 {
		// 如果JWT中没有用户ID，尝试从查询参数获取（用于测试）
		userIdStr := r.URL.Query().Get("user_id")
		if userIdStr != "" {
			if uid, err := strconv.ParseInt(userIdStr, 10, 64); err == nil {
				userId = uid
			}
		}
		
		if userId <= 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return nil
		}
	}

	l.Logger.Infof("WebSocket connection request from user %d", userId)
	
	// 使用统一的WebSocket管理器处理连接
	wsmanager.HandleWebSocket(w, r, userId)
	
	return nil
}
