package sms

import (
	"context"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// SecurityService 安全防护服务
type SecurityService struct {
	redis *redis.Redis
}

func NewSecurityService(redis *redis.Redis) *SecurityService {
	return &SecurityService{redis: redis}
}

// 检查是否允许发送
func (s *SecurityService) AllowSending(ctx context.Context, phone, ip string) error {
	// 检查手机号频率
	if err := s.checkRate(ctx, "phone:"+phone, 5, time.Minute*10); err != nil {
		return fmt.Errorf("手机号发送频率过高: %w", err)
	}

	// 检查IP频率
	if err := s.checkRate(ctx, "ip:"+ip, 20, time.Hour); err != nil {
		return fmt.Errorf("IP发送频率过高: %w", err)
	}

	// 检查手机号黑名单
	if s.isBlacklisted(ctx, "phone_blacklist", phone) {
		return errors.New("手机号在黑名单中")
	}

	// 检查IP黑名单
	if s.isBlacklisted(ctx, "ip_blacklist", ip) {
		return errors.New("IP在黑名单中")
	}

	return nil
}

// 记录发送行为
func (s *SecurityService) RecordSend(ctx context.Context, phone, ip string) {
	// 记录手机号发送
	s.recordSend(ctx, "phone:"+phone)

	// 记录IP发送
	s.recordSend(ctx, "ip:"+ip)
}

// 检查发送频率
func (s *SecurityService) checkRate(ctx context.Context, key string, maxCount int, window time.Duration) error {
	count, err := s.redis.GetCtx(ctx, key)
	if err != nil && err != redis.Nil {
		return fmt.Errorf("获取频率计数失败: %w", err)
	}

	if count != "" {
		cnt, _ := strconv.Atoi(count)
		if cnt >= maxCount {
			return errors.New("超过最大发送次数")
		}
	}
	return nil
}

// 记录发送行为
func (s *SecurityService) recordSend(ctx context.Context, key string) {
	// 使用管道提高性能
	err := s.redis.PipelinedCtx(ctx, func(pipe redis.Pipeliner) error {
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Duration(int(time.Hour.Seconds())))
		return nil
	})

	if err != nil {
		// 记录错误但不要中断主流程
		logx.WithContext(ctx).Errorf("记录发送行为失败: key:%s, err:%v", key, err)
	}
}

// 检查是否在黑名单中
func (s *SecurityService) isBlacklisted(ctx context.Context, listKey, value string) bool {
	exists, err := s.redis.SismemberCtx(ctx, listKey, value)
	return err == nil && exists
}

// 添加到黑名单
func (s *SecurityService) AddToBlacklist(ctx context.Context, listKey, value string, ttl time.Duration) error {
	_, err := s.redis.SaddCtx(ctx, listKey, value)
	if err != nil {
		return err
	}

	if ttl > 0 {
		err = s.redis.ExpireCtx(ctx, listKey, int(ttl.Seconds()))
	}
	return err
}
