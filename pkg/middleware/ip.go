package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
)

// ClientIPKey 是存储在上下文中的客户端 IP 键名
const ClientIPKey = "client_ip"

// ClientIPConfig 配置选项
type ClientIPConfig struct {
	// TrustedProxies 信任的代理 IP 列表（如负载均衡器、CDN）
	TrustedProxies []string

	// TrustedHeader 信任的 IP 头（默认 X-Forwarded-For）
	TrustedHeader string

	// UseRealIP 是否优先使用 X-Real-IP 头
	UseRealIP bool

	// Logger 日志记录器接口
	Logger interface {
		Errorf(format string, args ...interface{})
	}
}

// DefaultClientIPConfig 默认配置
var DefaultClientIPConfig = ClientIPConfig{
	TrustedProxies: []string{"127.0.0.1", "::1"}, // 默认信任本地
	TrustedHeader:  "X-Forwarded-For",
	UseRealIP:      false,
	Logger:         nil,
}

// ClientIPMiddleware 创建客户端 IP 中间件
func ClientIPMiddleware(config ...ClientIPConfig) func(http.Handler) http.Handler {
	var cfg ClientIPConfig
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = DefaultClientIPConfig
	}

	// 解析信任的代理 IP 为 CIDR 格式
	trustedCIDRs := make([]*net.IPNet, 0, len(cfg.TrustedProxies))
	for _, proxy := range cfg.TrustedProxies {
		if strings.Contains(proxy, "/") {
			if _, cidr, err := net.ParseCIDR(proxy); err == nil {
				trustedCIDRs = append(trustedCIDRs, cidr)
			} else if cfg.Logger != nil {
				cfg.Logger.Errorf("Failed to parse CIDR %s: %v", proxy, err)
			}
		} else {
			if ip := net.ParseIP(proxy); ip != nil {
				// 将单个 IP 转换为 CIDR（/32 或 /128）
				mask := net.CIDRMask(32, 32)
				if ip.To4() == nil {
					mask = net.CIDRMask(128, 128)
				}
				trustedCIDRs = append(trustedCIDRs, &net.IPNet{IP: ip, Mask: mask})
			} else if cfg.Logger != nil {
				cfg.Logger.Errorf("Failed to parse IP %s: %v", proxy, err)
			}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := extractClientIP(r, trustedCIDRs, cfg.TrustedHeader, cfg.UseRealIP)

			// 将客户端 IP 存入上下文
			ctx := context.WithValue(r.Context(), ClientIPKey, clientIP)

			// 继续处理请求
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractClientIP 从请求中提取客户端 IP
func extractClientIP(r *http.Request, trustedCIDRs []*net.IPNet, header string, useRealIP bool) string {
	// 1. 检查 X-Real-IP（如果启用）
	if useRealIP {
		if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			if isValidIP(realIP) {
				return realIP
			}
		}
	}

	// 2. 检查 X-Forwarded-For
	if xff := r.Header.Get(header); xff != "" {
		ips := strings.Split(xff, ",")
		for i := len(ips) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(ips[i])
			if ip == "" {
				continue
			}

			// 检查是否是信任的代理 IP
			if isTrustedProxy(ip, trustedCIDRs) {
				continue
			}

			// 第一个非信任代理的 IP 即为客户端 IP
			if isValidIP(ip) {
				return ip
			}
		}
	}

	// 3. 直接从 RemoteAddr 获取
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if isValidIP(ip) {
			return ip
		}
	}

	return "unknown"
}

// isTrustedProxy 检查 IP 是否是信任的代理
func isTrustedProxy(ipStr string, trustedCIDRs []*net.IPNet) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	for _, cidr := range trustedCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}

	return false
}

// isValidIP 检查是否是有效的 IP 地址
func isValidIP(ipStr string) bool {
	return net.ParseIP(ipStr) != nil
}

// GetClientIP 从上下文中获取客户端 IP
func GetClientIP(ctx context.Context) string {
	if ip, ok := ctx.Value(ClientIPKey).(string); ok {
		return ip
	}
	return "unknown"
}
