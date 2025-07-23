package sms

import (
	"encoding/json"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// 监控记录器接口
type MetricsRecorder interface {
	Record(result SendResult)
}

// 审计日志记录器接口
type AuditLogger interface {
	Log(result SendResult)
}

// 默认监控实现
type defaultMetricsRecorder struct{}

func NewDefaultMetricsRecorder() MetricsRecorder {
	return &defaultMetricsRecorder{}
}

func (r *defaultMetricsRecorder) Record(result SendResult) {
	// 实际项目中集成Prometheus等监控系统
	status := "success"
	if result.Error != nil {
		status = "failure"
	}

	logx.Infof("[Metrics] SMS发送 - 提供商:%s 状态:%s 耗时:%s",
		result.Provider, status, result.Cost)
}

// 默认审计实现
type defaultAuditLogger struct{}

func NewDefaultAuditLogger() AuditLogger {
	return &defaultAuditLogger{}
}

func (l *defaultAuditLogger) Log(result SendResult) {
	// 手机号脱敏处理
	phone := DesensitizePhone(result.Phone)

	// 记录审计日志
	auditLog := struct {
		Type       string `json:"type"`
		Provider   string `json:"provider"`
		Phone      string `json:"phone"`
		TemplateID string `json:"template_id"`
		Success    bool   `json:"success"`
		Error      string `json:"error,omitempty"`
		Cost       string `json:"cost"`
		Timestamp  int64  `json:"timestamp"`
	}{
		Type:       "sms_send",
		Provider:   string(result.Provider),
		Phone:      phone,
		TemplateID: result.TemplateID,
		Success:    result.Success,
		Cost:       result.Cost.String(),
		Timestamp:  time.Now().Unix(),
	}

	if result.Error != nil {
		auditLog.Error = result.Error.Error()
	}

	logJson, _ := json.Marshal(auditLog)
	logx.Infof(string(logJson))
}

// DesensitizePhone 手机号脱敏
func DesensitizePhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}
