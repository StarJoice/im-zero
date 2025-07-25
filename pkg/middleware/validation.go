package middleware

import (
	"net/http"
	"strings"

	"im-zero/pkg/tool"
	"im-zero/pkg/xerrs"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// ValidationMiddleware 输入验证中间件
func ValidationMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 验证请求头
			if err := validateHeaders(r); err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}

			// 验证请求体大小
			if err := validateRequestSize(r); err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}

			// 验证Content-Type（对于POST/PUT请求）
			if err := validateContentType(r); err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// validateHeaders 验证请求头
func validateHeaders(r *http.Request) error {
	// 验证User-Agent
	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		return xerrs.NewErrCodeMsg(xerrs.INVALID_REQUEST, "User-Agent header is required")
	}
	if len(userAgent) > 500 {
		return xerrs.NewErrCodeMsg(xerrs.INVALID_REQUEST, "User-Agent header too long")
	}

	// 验证X-Forwarded-For
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" && len(xff) > 200 {
		return xerrs.NewErrCodeMsg(xerrs.INVALID_REQUEST, "X-Forwarded-For header too long")
	}

	return nil
}

// validateRequestSize 验证请求体大小
func validateRequestSize(r *http.Request) error {
	const maxRequestSize = 10 << 20 // 10MB

	if r.ContentLength > maxRequestSize {
		return xerrs.NewErrCodeMsg(xerrs.INVALID_REQUEST, "request body too large")
	}

	return nil
}

// validateContentType 验证Content-Type
func validateContentType(r *http.Request) error {
	if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
		contentType := r.Header.Get("Content-Type")
		if contentType == "" {
			return xerrs.NewErrCodeMsg(xerrs.INVALID_REQUEST, "Content-Type header is required")
		}

		// 检查是否为支持的Content-Type
		supportedTypes := []string{
			"application/json",
			"application/x-www-form-urlencoded",
			"multipart/form-data",
		}

		isSupported := false
		for _, supportedType := range supportedTypes {
			if strings.Contains(contentType, supportedType) {
				isSupported = true
				break
			}
		}

		if !isSupported {
			return xerrs.NewErrCodeMsg(xerrs.INVALID_REQUEST, "unsupported Content-Type")
		}
	}

	return nil
}

// MobileValidationMiddleware 手机号验证中间件
func MobileValidationMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 只对包含mobile参数的请求进行验证
			if mobile := extractMobileFromRequest(r); mobile != "" {
				if !tool.ValidateMobile(mobile) {
					httpx.ErrorCtx(r.Context(), w, xerrs.NewErrCodeMsg(xerrs.INVALID_MOBILE, "invalid mobile format"))
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractMobileFromRequest 从请求中提取手机号
func extractMobileFromRequest(r *http.Request) string {
	// 从URL参数中提取
	if mobile := r.URL.Query().Get("mobile"); mobile != "" {
		return mobile
	}

	// 从表单中提取
	if err := r.ParseForm(); err == nil {
		if mobile := r.FormValue("mobile"); mobile != "" {
			return mobile
		}
	}

	return ""
}

// RateLimitMiddleware 简单的频率限制中间件
func RateLimitMiddleware(maxRequests int) func(http.Handler) http.Handler {
	requestCounts := make(map[string]int)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			
			// 简单的内存限流（生产环境应使用Redis）
			requestCounts[clientIP]++
			if requestCounts[clientIP] > maxRequests {
				httpx.ErrorCtx(r.Context(), w, xerrs.NewErrCodeMsg(xerrs.RATE_LIMIT_ERROR, "too many requests"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP 获取客户端IP
func getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 检查X-Real-IP头  
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// 使用RemoteAddr
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}

	return ip
}

// SecurityHeadersMiddleware 安全头中间件
func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 设置安全响应头
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")

			next.ServeHTTP(w, r)
		})
	}
}