package message

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"im-zero/app/message/cmd/api/internal/logic/message"
	"im-zero/app/message/cmd/api/internal/svc"
	"im-zero/app/message/cmd/api/internal/types"
)

// 获取聊天记录
func GetChatHistoryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetChatHistoryReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := message.NewGetChatHistoryLogic(r.Context(), svcCtx)
		resp, err := l.GetChatHistory(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
