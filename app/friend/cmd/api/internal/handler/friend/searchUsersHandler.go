package friend

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"im-zero/app/friend/cmd/api/internal/logic/friend"
	"im-zero/app/friend/cmd/api/internal/svc"
	"im-zero/app/friend/cmd/api/internal/types"
)

// 搜索用户
func SearchUsersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SearchUsersReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := friend.NewSearchUsersLogic(r.Context(), svcCtx)
		resp, err := l.SearchUsers(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
