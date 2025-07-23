package verifycode

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"im-zero/app/verifycode/cmd/api/internal/logic/verifycode"
	"im-zero/app/verifycode/cmd/api/internal/svc"
	"im-zero/app/verifycode/cmd/api/internal/types"
)

// send verifycode
func SendSmsCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SendSmsCodeReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := verifycode.NewSendSmsCodeLogic(r.Context(), svcCtx)
		resp, err := l.SendSmsCode(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
