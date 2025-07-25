package group

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"im-zero/app/group/cmd/api/internal/logic/group"
	"im-zero/app/group/cmd/api/internal/svc"
	"im-zero/app/group/cmd/api/internal/types"
)

// 邀请用户入群
func InviteUsersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.InviteUsersReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := group.NewInviteUsersLogic(r.Context(), svcCtx)
		resp, err := l.InviteUsers(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
