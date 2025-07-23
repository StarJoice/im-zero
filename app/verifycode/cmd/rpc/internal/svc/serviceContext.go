package svc

import (
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"im-zero/app/verifycode/cmd/rpc/internal/config"
	"im-zero/pkg/sms"
)

type ServiceContext struct {
	Config       config.Config
	Redis        *redis.Redis
	SmsService   sms.Service
	SmsSecurity  *sms.SecurityService
	SmsTemplates *sms.TemplateManager
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 1. 创建Redis客户端
	redisClient := redis.MustNewRedis(c.Redis.RedisConf)

	// 2. 创建短信服务
	var smsService sms.Service
	if c.Sms.Provider != "" {
		var err error
		smsService, err = sms.NewService(sms.Config{
			Provider:       sms.Provider(c.Sms.Provider),
			AccessKey:      c.Sms.AccessKey,
			AccessSecret:   c.Sms.AccessSecret,
			DefaultSign:    c.Sms.SignName,
			DefaultTplCode: c.Sms.TemplateCode,
		},
			sms.NewDefaultMetricsRecorder(),
			sms.NewDefaultAuditLogger())

		if err != nil {
			logx.Must(fmt.Errorf("短信服务初始化失败: %w", err))
		}
	} else {
		// 使用模拟服务作为回退
		smsService = sms.NewMockService()
		logx.Infof("未配置短信服务商，使用模拟短信服务")
	}

	// 3. 创建安全防护服务
	securityService := sms.NewSecurityService(redisClient)

	// 4. 创建模板管理器并加载模板
	templateManager := sms.NewTemplateManager()
	if err := loadSmsTemplates(templateManager, c.Sms.Templates); err != nil {
		logx.Must(fmt.Errorf("加载短信模板失败: %w", err))
	}

	// 5. 添加安全防护中间件
	smsService = sms.NewSecurityMiddleware(smsService, securityService)

	return &ServiceContext{
		Config:       c,
		Redis:        redisClient,
		SmsService:   smsService,
		SmsSecurity:  securityService,
		SmsTemplates: templateManager,
	}
}

// 加载短信模板
func loadSmsTemplates(manager *sms.TemplateManager, templates []config.SmsTemplate) error {
	if len(templates) == 0 {
		return errors.New("未配置短信模板")
	}

	for _, tpl := range templates {
		// 验证模板配置
		if tpl.ID == "" {
			return errors.New("短信模板ID不能为空")
		}
		if len(tpl.Params) == 0 {
			return fmt.Errorf("模板[%s]缺少参数配置", tpl.ID)
		}

		manager.AddTemplate(sms.Template{
			ID:          tpl.ID,
			Provider:    sms.Provider(tpl.Provider),
			Description: tpl.Description,
			Content:     tpl.Content,
			Params:      tpl.Params,
			Enabled:     tpl.Enabled,
			RateLimit:   tpl.RateLimit,
		})

		logx.Infof("加载短信模板: ID=%s, 描述=%s", tpl.ID, tpl.Description)
	}
	return nil
}
