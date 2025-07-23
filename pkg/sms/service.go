package sms

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// 短信服务接口
type Service interface {
	Send(ctx context.Context, phone, templateID string, params map[string]string) error
}

// 短信服务提供商
type Provider string

const (
	ProviderAliyun  Provider = "aliyun"
	ProviderTencent Provider = "tencent"
	ProviderMock    Provider = "mock"
)

// 短信服务配置
type Config struct {
	Provider       Provider
	AccessKey      string
	AccessSecret   string
	DefaultSign    string
	DefaultTplCode string
}

// 短信发送结果
type SendResult struct {
	Success    bool
	Provider   Provider
	Phone      string
	TemplateID string
	Cost       time.Duration
	Error      error
}

// 带监控和审计的短信服务
type monitoredService struct {
	provider     Service
	providerName Provider
	metrics      MetricsRecorder
	audit        AuditLogger
}

func (s *monitoredService) Send(ctx context.Context, phone, templateID string, params map[string]string) error {
	start := time.Now()
	err := s.provider.Send(ctx, phone, templateID, params)
	duration := time.Since(start)

	result := SendResult{
		Success:    err == nil,
		Provider:   s.providerName,
		Phone:      phone,
		TemplateID: templateID,
		Cost:       duration,
		Error:      err,
	}

	// 记录监控指标
	s.metrics.Record(result)

	// 记录审计日志
	s.audit.Log(result)

	return err
}

// 创建短信服务
func NewService(cfg Config, metrics MetricsRecorder, audit AuditLogger) (Service, error) {
	var provider Service
	var err error

	switch cfg.Provider {
	case ProviderAliyun:
		provider, err = NewAliyunService(cfg.AccessKey, cfg.AccessSecret)
	case ProviderTencent:
		provider, err = NewTencentService(cfg.AccessKey, cfg.AccessSecret)
	case ProviderMock:
		provider = NewMockService()
	default:
		return nil, errors.New("unsupported sms provider")
	}

	if err != nil {
		return nil, err
	}

	return &monitoredService{
		provider:     provider,
		providerName: cfg.Provider,
		metrics:      metrics,
		audit:        audit,
	}, nil
}

// 阿里云短信服务实现
type aliyunService struct {
	accessKey    string
	accessSecret string
}

func NewAliyunService(accessKey, accessSecret string) (Service, error) {
	if accessKey == "" || accessSecret == "" {
		return nil, errors.New("aliyun config missing")
	}
	return &aliyunService{
		accessKey:    accessKey,
		accessSecret: accessSecret,
	}, nil
}

func (s *aliyunService) Send(ctx context.Context, phone, templateID string, params map[string]string) error {
	// 实际调用阿里云SDK
	logx.WithContext(ctx).Infof("阿里云短信发送: phone=%s, tpl=%s, params=%v",
		phone, templateID, params)

	return nil
}

// 腾讯云短信服务实现
type tencentService struct {
	accessKey    string
	accessSecret string
}

func NewTencentService(accessKey, accessSecret string) (Service, error) {
	if accessKey == "" || accessSecret == "" {
		return nil, errors.New("tencent config missing")
	}
	return &tencentService{
		accessKey:    accessKey,
		accessSecret: accessSecret,
	}, nil
}

func (s *tencentService) Send(ctx context.Context, phone, templateID string, params map[string]string) error {
	// todo 实际调用腾讯云SDK
	logx.WithContext(ctx).Infof("腾讯云短信发送: phone=%s, tpl=%s, params=%v",
		phone, templateID, params)

	// 模拟成功
	return nil
}

// 模拟短信服务（用于开发和测试）
type mockService struct {
	sentMessages map[string]string
	mu           sync.Mutex
}

func NewMockService() Service {
	return &mockService{
		sentMessages: make(map[string]string),
	}
}

func (s *mockService) Send(ctx context.Context, phone, templateID string, params map[string]string) error {
	logx.WithContext(ctx).Infof("模拟短信发送: phone=%s, tpl=%s, params=%v",
		phone, templateID, params)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.sentMessages[phone] = fmt.Sprintf("模板[%s]参数:%v", templateID, params)
	return nil
}

func (s *mockService) GetSentMessage(phone string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sentMessages[phone]
}
