package sms

import (
	"context"
	"strings"
)

// 安全中间件
type securityMiddleware struct {
	next   Service
	secSvc *SecurityService
}

func NewSecurityMiddleware(next Service, secSvc *SecurityService) Service {
	return &securityMiddleware{
		next:   next,
		secSvc: secSvc,
	}
}

func (m *securityMiddleware) Send(ctx context.Context, phone, templateID string, params map[string]string) error {
	// 获取客户端IP
	ip := extractClientIP(ctx)

	// 执行安全检查
	if err := m.secSvc.AllowSending(ctx, phone, ip); err != nil {
		// 根据错误类型细化处理
		if strings.Contains(err.Error(), "手机号发送频率过高") {
			return NewSmsError(ErrCodeRateLimit, "发送频率过高")
		}

		if strings.Contains(err.Error(), "IP发送频率过高") {
			return NewSmsError(ErrCodeIpRateLimit, "当前网络环境操作频繁")
		}

		if strings.Contains(err.Error(), "黑名单") {
			return NewSmsError(ErrCodeBlacklisted, "操作受限")
		}

		return NewSmsError(ErrCodeSecurityCheck, "安全验证失败")
	}

	// 调用实际服务
	err := m.next.Send(ctx, phone, templateID, params)
	if err != nil {
		return err
	}

	// 记录发送行为
	m.secSvc.RecordSend(ctx, phone, ip)
	return nil
}

// 提取客户端IP
func extractClientIP(ctx context.Context) string {
	// 实际项目中从上下文获取客户端IP
	if ip, ok := ctx.Value("client_ip").(string); ok {
		return ip
	}
	return "127.0.0.1" // 默认值
}

// 短信错误定义
type SmsError struct {
	Code    ErrCode
	Message string
}

func (e *SmsError) Error() string {
	return e.Message
}

func NewSmsError(code ErrCode, msg string) *SmsError {
	return &SmsError{
		Code:    code,
		Message: msg,
	}
}

type ErrCode int

const (
	ErrCodeRateLimit ErrCode = iota + 1000
	ErrCodeIpRateLimit
	ErrCodeBlacklisted
	ErrCodeSecurityCheck
	ErrCodeTemplateInvalid
	ErrCodeProviderFailure
)
