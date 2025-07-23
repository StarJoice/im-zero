package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Sms struct {
		Provider     string        `json:",default=mock"` // aliyun, tencent, mock
		AccessKey    string        `json:",optional"`     // 访问密钥ID
		AccessSecret string        `json:",optional"`     // 访问密钥
		SignName     string        `json:",optional"`     // 短信签名
		TemplateCode string        `json:",optional"`     // 默认模板CODE
		Templates    []SmsTemplate `json:",optional"`     // 新增：模板配置
	} `json:",optional"`
}

// 新增短信模板配置结构体
type SmsTemplate struct {
	ID          string   `json:"id"`          // 模板ID
	Provider    string   `json:"provider"`    // 服务商
	Description string   `json:"description"` // 模板描述
	Content     string   `json:"content"`     // 模板内容
	Params      []string `json:"params"`      // 模板参数
	Enabled     bool     `json:"enabled"`     // 是否启用
	RateLimit   int      `json:"RateLimit"`   // 频率限制（次/分钟）
}
