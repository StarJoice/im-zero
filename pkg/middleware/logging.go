package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// LoggingMiddleware 综合日志中间件
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 创建响应记录器
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     200,
				body:          &bytes.Buffer{},
			}

			// 记录请求信息
			logRequestInfo(r, start)

			// 执行请求
			next.ServeHTTP(recorder, r)

			// 记录响应信息
			duration := time.Since(start)
			logResponseInfo(r, recorder, duration)
		})
	}
}

// responseRecorder 响应记录器
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	r.body.Write(data)
	return r.ResponseWriter.Write(data)
}

// logRequestInfo 记录请求信息
func logRequestInfo(r *http.Request, start time.Time) {
	// 获取客户端IP
	clientIP := getClientIP(r)
	
	// 获取用户代理
	userAgent := r.Header.Get("User-Agent")
	
	// 记录请求基本信息
	logx.WithContext(r.Context()).Infow("HTTP Request Started",
		logx.Field("method", r.Method),
		logx.Field("url", r.URL.String()),
		logx.Field("path", r.URL.Path),
		logx.Field("query", r.URL.RawQuery),
		logx.Field("remote_addr", r.RemoteAddr),
		logx.Field("client_ip", clientIP),
		logx.Field("user_agent", truncateString(userAgent, 200)),
		logx.Field("content_length", r.ContentLength),
		logx.Field("start_time", start.Format(time.RFC3339Nano)),
	)

	// 记录重要的请求头
	logImportantHeaders(r)

	// 记录请求体（仅对特定路径且非敏感信息）
	logRequestBody(r)
}

// logResponseInfo 记录响应信息
func logResponseInfo(r *http.Request, recorder *responseRecorder, duration time.Duration) {
	// 计算响应状态级别
	level := getLogLevel(recorder.statusCode)
	
	fields := []logx.LogField{
		logx.Field("method", r.Method),
		logx.Field("path", r.URL.Path),
		logx.Field("status_code", recorder.statusCode),
		logx.Field("duration_ms", duration.Milliseconds()),
		logx.Field("response_size", recorder.body.Len()),
	}

	message := "HTTP Request Completed"
	
	switch level {
	case "info":
		logx.WithContext(r.Context()).Infow(message, fields...)
	case "warn":
		logx.WithContext(r.Context()).Sloww(message, fields...)
	case "error":
		logx.WithContext(r.Context()).Errorw(message, fields...)
		// 对于错误状态，记录响应体
		if recorder.body.Len() > 0 && recorder.body.Len() < 1000 {
			logx.WithContext(r.Context()).Errorw("Error Response Body",
				logx.Field("body", recorder.body.String()),
			)
		}
	}

	// 记录慢请求
	if duration > 2*time.Second {
		logx.WithContext(r.Context()).Sloww("Slow Request Detected",
			logx.Field("method", r.Method),
			logx.Field("path", r.URL.Path),
			logx.Field("duration_ms", duration.Milliseconds()),
		)
	}
}

// logImportantHeaders 记录重要的请求头
func logImportantHeaders(r *http.Request) {
	importantHeaders := map[string]string{
		"Authorization":    r.Header.Get("Authorization"),
		"Content-Type":     r.Header.Get("Content-Type"),
		"Accept":          r.Header.Get("Accept"),
		"X-Request-ID":    r.Header.Get("X-Request-ID"),
		"X-Forwarded-For": r.Header.Get("X-Forwarded-For"),
		"Referer":         r.Header.Get("Referer"),
	}

	var headerFields []logx.LogField
	for key, value := range importantHeaders {
		if value != "" {
			// 敏感信息脱敏
			if key == "Authorization" && strings.HasPrefix(value, "Bearer ") {
				value = "Bearer ***"
			}
			headerFields = append(headerFields, logx.Field("header_"+strings.ToLower(key), truncateString(value, 100)))
		}
	}

	if len(headerFields) > 0 {
		logx.WithContext(r.Context()).Infow("Request Headers", headerFields...)
	}
}

// logRequestBody 记录请求体（非敏感路径）
func logRequestBody(r *http.Request) {
	// 只对特定路径记录请求体
	sensititivePaths := []string{
		"/api/user/login",
		"/api/user/register", 
		"/api/verifycode/send",
	}

	shouldLog := true
	for _, path := range sensititivePaths {
		if strings.Contains(r.URL.Path, path) {
			shouldLog = false
			break
		}
	}

	if shouldLog && r.Method == "POST" && r.ContentLength > 0 && r.ContentLength < 1000 {
		body, err := io.ReadAll(r.Body)
		if err == nil {
			// 重置请求体
			r.Body = io.NopCloser(bytes.NewBuffer(body))
			
			logx.WithContext(r.Context()).Infow("Request Body",
				logx.Field("body", string(body)),
			)
		}
	}
}

// getLogLevel 根据状态码获取日志级别
func getLogLevel(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "error"
	case statusCode >= 400:
		return "warn"
	default:
		return "info"
	}
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// PerformanceMiddleware 性能监控中间件
func PerformanceMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			next.ServeHTTP(w, r)
			
			duration := time.Since(start)
			
			// 记录性能指标
			logx.WithContext(r.Context()).Infow("Performance Metrics",
				logx.Field("method", r.Method),
				logx.Field("path", r.URL.Path),
				logx.Field("duration_ms", duration.Milliseconds()),
				logx.Field("duration_ns", duration.Nanoseconds()),
			)
			
			// 性能预警
			if duration > 5*time.Second {
				logx.WithContext(r.Context()).Errorw("Performance Alert: Very Slow Request",
					logx.Field("method", r.Method),
					logx.Field("path", r.URL.Path),
					logx.Field("duration_s", duration.Seconds()),
				)
			} else if duration > 1*time.Second {
				logx.WithContext(r.Context()).Sloww("Performance Warning: Slow Request",
					logx.Field("method", r.Method),
					logx.Field("path", r.URL.Path),
					logx.Field("duration_ms", duration.Milliseconds()),
				)
			}
		})
	}
}