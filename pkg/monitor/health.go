package monitor

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// HealthChecker 系统健康检查器
type HealthChecker struct {
	checks map[string]HealthCheck
	mutex  sync.RWMutex
}

// HealthCheck 健康检查接口
type HealthCheck interface {
	Name() string
	Check(ctx context.Context) error
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp int64                  `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    int64                  `json:"uptime"`
	Checks    map[string]CheckResult `json:"checks"`
	System    SystemInfo             `json:"system"`
}

// CheckResult 检查结果
type CheckResult struct {
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	Duration  int64  `json:"duration_ms"`
	Timestamp int64  `json:"timestamp"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
	MemStats     MemoryStats `json:"memory"`
}

// MemoryStats 内存统计
type MemoryStats struct {
	Alloc      uint64 `json:"alloc_mb"`
	TotalAlloc uint64 `json:"total_alloc_mb"`
	Sys        uint64 `json:"sys_mb"`
	NumGC      uint32 `json:"num_gc"`
}

var (
	globalHealthChecker *HealthChecker
	once                sync.Once
	startTime           = time.Now()
)

// GetHealthChecker 获取全局健康检查器实例
func GetHealthChecker() *HealthChecker {
	once.Do(func() {
		globalHealthChecker = &HealthChecker{
			checks: make(map[string]HealthCheck),
		}
	})
	return globalHealthChecker
}

// RegisterCheck 注册健康检查
func (h *HealthChecker) RegisterCheck(check HealthCheck) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.checks[check.Name()] = check
}

// CheckHealth 执行健康检查
func (h *HealthChecker) CheckHealth(ctx context.Context) *HealthStatus {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	status := &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
		Version:   "1.0.0", // 可以从配置或构建时注入
		Uptime:    int64(time.Since(startTime).Seconds()),
		Checks:    make(map[string]CheckResult),
		System:    h.getSystemInfo(),
	}

	// 执行各项检查
	for name, check := range h.checks {
		result := h.executeCheck(ctx, check)
		status.Checks[name] = result
		
		// 如果有检查失败，整体状态标记为不健康
		if result.Status != "healthy" {
			status.Status = "unhealthy"
		}
	}

	return status
}

// executeCheck 执行单个检查
func (h *HealthChecker) executeCheck(ctx context.Context, check HealthCheck) CheckResult {
	start := time.Now()
	
	// 设置检查超时
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	err := check.Check(checkCtx)
	duration := time.Since(start)
	
	result := CheckResult{
		Duration:  duration.Milliseconds(),
		Timestamp: time.Now().Unix(),
	}
	
	if err != nil {
		result.Status = "unhealthy"
		result.Message = err.Error()
		logx.WithContext(ctx).Errorw("Health check failed",
			logx.Field("check", check.Name()),
			logx.Field("error", err.Error()),
			logx.Field("duration_ms", duration.Milliseconds()),
		)
	} else {
		result.Status = "healthy"
		logx.WithContext(ctx).Infow("Health check passed",
			logx.Field("check", check.Name()),
			logx.Field("duration_ms", duration.Milliseconds()),
		)
	}
	
	return result
}

// getSystemInfo 获取系统信息
func (h *HealthChecker) getSystemInfo() SystemInfo {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	return SystemInfo{
		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		MemStats: MemoryStats{
			Alloc:      memStats.Alloc / 1024 / 1024,
			TotalAlloc: memStats.TotalAlloc / 1024 / 1024,
			Sys:        memStats.Sys / 1024 / 1024,
			NumGC:      memStats.NumGC,
		},
	}
}

// DatabaseCheck 数据库健康检查
type DatabaseCheck struct {
	name string
	db   *sql.DB
}

func NewDatabaseCheck(name string, db *sql.DB) *DatabaseCheck {
	return &DatabaseCheck{
		name: name,
		db:   db,
	}
}

func (d *DatabaseCheck) Name() string {
	return d.name
}

func (d *DatabaseCheck) Check(ctx context.Context) error {
	if d.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	// 执行简单的ping检查
	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	
	// 检查连接池状态
	stats := d.db.Stats()
	if stats.OpenConnections == 0 {
		return fmt.Errorf("no open database connections")
	}
	
	return nil
}

// RedisCheck Redis健康检查
type RedisCheck struct {
	name   string
	client RedisClient
}

type RedisClient interface {
	PingCtx(ctx context.Context) error
}

func NewRedisCheck(name string, client RedisClient) *RedisCheck {
	return &RedisCheck{
		name:   name,
		client: client,
	}
}

func (r *RedisCheck) Name() string {
	return r.name
}

func (r *RedisCheck) Check(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("redis client is nil")
	}
	
	return r.client.PingCtx(ctx)
}

// HTTPServiceCheck HTTP服务健康检查
type HTTPServiceCheck struct {
	name string
	url  string
}

func NewHTTPServiceCheck(name, url string) *HTTPServiceCheck {
	return &HTTPServiceCheck{
		name: name,
		url:  url,
	}
}

func (h *HTTPServiceCheck) Name() string {
	return h.name
}

func (h *HTTPServiceCheck) Check(ctx context.Context) error {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", h.url, nil)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("http status code: %d", resp.StatusCode)
	}
	
	return nil
}

// HealthCheckHandler HTTP健康检查处理器
func HealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checker := GetHealthChecker()
		status := checker.CheckHealth(r.Context())
		
		// 根据健康状态设置HTTP状态码
		if status.Status == "healthy" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		
		httpx.OkJsonCtx(r.Context(), w, status)
	}
}

// StartPeriodicHealthCheck 启动定期健康检查
func StartPeriodicHealthCheck(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				checker := GetHealthChecker()
				status := checker.CheckHealth(context.Background())
				
				if status.Status != "healthy" {
					logx.Errorw("System health check failed",
						logx.Field("status", status.Status),
						logx.Field("failed_checks", getFailedChecks(status.Checks)),
					)
				} else {
					logx.Infow("System health check passed",
						logx.Field("uptime_seconds", status.Uptime),
						logx.Field("memory_mb", status.System.MemStats.Alloc),
						logx.Field("goroutines", status.System.NumGoroutine),
					)
				}
			}
		}
	}()
}

// getFailedChecks 获取失败的检查项
func getFailedChecks(checks map[string]CheckResult) []string {
	var failed []string
	for name, result := range checks {
		if result.Status != "healthy" {
			failed = append(failed, name)
		}
	}
	return failed
}