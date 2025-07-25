package ws

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"im-zero/app/message/cmd/api/internal/logic/ws"
	"im-zero/app/message/cmd/api/internal/svc"
	"im-zero/app/message/cmd/api/internal/types"
)

// WebSocket连接
func WsConnectHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WsConnectReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := ws.NewWsConnectLogic(r.Context(), svcCtx)
		err := l.WsConnect(&req, w, r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		}
		// WebSocket连接不需要返回OK响应，因为连接已经被升级
	}
}
