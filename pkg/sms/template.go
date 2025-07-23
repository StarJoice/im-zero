package sms

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"sync"
)

// Template 短信模板
type Template struct {
	ID          string   `json:"id"`          // 模板ID
	Provider    Provider `json:"provider"`    // 服务商
	Description string   `json:"description"` // 模板描述
	Content     string   `json:"content"`     // 模板内容
	Params      []string `json:"params"`      // 模板参数
	Enabled     bool     `json:"enabled"`     // 是否启用
	RateLimit   int      `json:"rate_limit"`  // 频率限制（次/分钟）
}

// TemplateManager 模板管理器
type TemplateManager struct {
	templates map[string]Template // 模板ID -> Template
	mu        sync.RWMutex
}

func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		templates: make(map[string]Template),
	}
}

// 添加或更新模板
func (m *TemplateManager) AddTemplate(tpl Template) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.templates[tpl.ID] = tpl
}

// 获取模板
func (m *TemplateManager) GetTemplate(id string) (Template, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	tpl, exists := m.templates[id]
	return tpl, exists
}

// ValidateParams 验证模板参数
func (m *TemplateManager) ValidateParams(tplID string, params map[string]string) error {
	tpl, exists := m.GetTemplate(tplID)
	if !exists {
		return errors.New("模板不存在")
	}

	// 检查参数数量
	if len(params) != len(tpl.Params) {
		return errors.New("参数数量不匹配")
	}

	// 检查参数是否存在
	for _, param := range tpl.Params {
		if _, ok := params[param]; !ok {
			return fmt.Errorf("缺少参数: %s", param)
		}
	}

	return nil
}

// 从JSON加载模板
func (m *TemplateManager) LoadFromJSON(jsonStr string) error {
	var templates []Template
	if err := json.Unmarshal([]byte(jsonStr), &templates); err != nil {
		return err
	}

	for _, tpl := range templates {
		m.AddTemplate(tpl)
	}
	return nil
}
