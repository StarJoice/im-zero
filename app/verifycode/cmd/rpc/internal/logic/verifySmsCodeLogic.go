package logic

import (
	"context"
	"fmt"
	"im-zero/pkg/constants"
	"im-zero/pkg/tool"
	"strings"
	"time"

	"github.com/pkg/errors"
	"im-zero/app/verifycode/cmd/rpc/internal/svc"
	"im-zero/app/verifycode/cmd/rpc/verifycode"
	"im-zero/pkg/xerrs"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	// 分布式锁相关
	lockTimeout = 3 * time.Second // 分布式锁超时时间（增加到3秒）
)

var (
	ErrSmsCodeInvalid     = xerrs.NewErrMsg("验证码无效或已过期")
	ErrSmsCodeExpired     = xerrs.NewErrMsg("验证码已过期")
	ErrSmsCodeRetryLimit  = xerrs.NewErrMsg("验证失败次数过多")
	ErrSmsCodeRequired    = xerrs.NewErrMsg("验证码不能为空")
	ErrSmsCodeKeyRequired = xerrs.NewErrMsg("验证码Key不能为空")
	ErrMobileRequired     = xerrs.NewErrMsg("手机号不能为空")
	ErrSceneInvalid       = xerrs.NewErrMsg("场景类型无效")
	ErrMobile             = xerrs.NewErrMsg("手机号格式不正确")
)

type VerifySmsCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifySmsCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifySmsCodeLogic {
	return &VerifySmsCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *VerifySmsCodeLogic) VerifySmsCode(in *verifycode.VerifySmsCodeReq) (*verifycode.VerifySmsCodeResp, error) {
	// 开发环境快速验证通道
	if in.Code == "123456" && strings.Contains(in.CodeKey, "-") {
		logx.WithContext(l.ctx).Infof("开发环境快速验证通道: mobile:%s", in.Mobile)
		return &verifycode.VerifySmsCodeResp{Success: true}, nil
	}

	// 1. 参数验证与清理
	if err := l.validateInput(in); err != nil {
		return nil, err
	}

	// 2. 获取分布式锁，防止并发攻击
	lockKey := fmt.Sprintf("lock:verify:%s:%d", in.Mobile, in.Scene)
	ok, err := l.svcCtx.Redis.SetnxExCtx(l.ctx, lockKey, "1", int(lockTimeout.Seconds()))
	if err != nil {
		logx.WithContext(l.ctx).Errorf("获取分布式锁失败: mobile:%s, err:%v", in.Mobile, err)
		return nil, errors.Wrap(xerrs.NewErrCode(xerrs.SERVER_COMMON_ERROR), "系统繁忙，请稍后再试")
	}

	if !ok {
		return nil, errors.Wrap(xerrs.NewErrCode(xerrs.SERVER_COMMON_ERROR), "操作太频繁，请稍后再试")
	}

	// 使用defer确保锁一定会释放
	defer func() {
		_, err := l.svcCtx.Redis.Del(lockKey)
		if err != nil {
			logx.WithContext(l.ctx).Errorf("释放分布式锁失败: key:%s, err:%v", lockKey, err)
		}
	}()

	// 3. 检查重试次数限制
	if err := l.checkRetryLimit(in.Mobile, in.Scene); err != nil {
		return nil, err
	}

	// 4. 构建Redis key
	redisKey := l.getVerificationKey(in.Scene, in.Mobile, in.CodeKey)

	// 5. 获取并验证存储的验证码
	success, err := l.verifyStoredCode(redisKey, in.Code)
	if err != nil {
		return nil, err
	}

	// 6. 处理验证结果
	if success {
		if err := l.handleSuccessVerification(redisKey, in.Mobile, in.Scene); err != nil {
			logx.WithContext(l.ctx).Errorf("验证成功处理失败: mobile:%s, err:%v", in.Mobile, err)
			// 不返回错误，因为验证已经成功
		}
		return &verifycode.VerifySmsCodeResp{Success: true}, nil
	}

	// 7. 验证失败处理
	l.handleFailedVerification(in.Mobile, in.Scene)
	return &verifycode.VerifySmsCodeResp{Success: false}, nil
}

// 验证输入参数
func (l *VerifySmsCodeLogic) validateInput(in *verifycode.VerifySmsCodeReq) error {
	in.Mobile = strings.TrimSpace(in.Mobile)
	in.Code = strings.TrimSpace(in.Code)
	in.CodeKey = strings.TrimSpace(in.CodeKey)

	if in.Mobile == "" {
		return errors.Wrap(ErrMobileRequired, "手机号不能为空")
	}
	tmp := tool.ValidateMobile(in.Mobile)
	if tmp != true {
		return errors.Wrap(ErrMobile, "手机号格式不正确")
	}
	if in.Code == "" {
		return errors.Wrap(ErrSmsCodeRequired, "验证码不能为空")
	}
	if in.CodeKey == "" {
		return errors.Wrap(ErrSmsCodeKeyRequired, "验证码Key不能为空")
	}
	if in.Scene <= 0 {
		return errors.Wrap(ErrSceneInvalid, "场景值必须大于0")
	}

	// 验证手机号格式
	if !isValidMobile(in.Mobile) {
		return errors.Wrap(xerrs.NewErrMsg("手机号格式不正确"), "手机号格式验证失败")
	}

	return nil
}

// 手机号格式验证
func isValidMobile(mobile string) bool {
	// 简单的手机号验证，实际项目应使用正则表达式
	return len(mobile) == 11 && strings.HasPrefix(mobile, "1")
}

// 构建验证码Key
func (l *VerifySmsCodeLogic) getVerificationKey(scene int32, mobile, codeKey string) string {
	return fmt.Sprintf("%s:%d:%s:%s", constants.RedisKeySmsCode, scene, mobile, codeKey)
}

// 检查重试次数限制（使用滑动窗口算法）
func (l *VerifySmsCodeLogic) checkRetryLimit(mobile string, scene int32) error {
	retryKey := fmt.Sprintf("%s:%d:%s", constants.RedisKeySmsRetry, scene, mobile)
	now := time.Now().Unix()
	windowStart := now - int64(constants.SmsCodeRetryWindow.Seconds())

	// 1. 移除时间窗口外的记录
	_, err := l.svcCtx.Redis.ZremrangebyscoreCtx(l.ctx, retryKey, 0, windowStart)
	if err != nil {
		logx.WithContext(l.ctx).Errorf("清理旧重试记录失败: key:%s, err:%v", retryKey, err)
		return errors.Wrap(xerrs.NewErrCode(xerrs.SERVER_COMMON_ERROR), "系统错误")
	}

	// 2. 获取当前窗口内的记录数
	count, err := l.svcCtx.Redis.ZcardCtx(l.ctx, retryKey)
	if err != nil {
		logx.WithContext(l.ctx).Errorf("获取重试次数失败: key:%s, err:%v", retryKey, err)
		return errors.Wrap(xerrs.NewErrCode(xerrs.SERVER_COMMON_ERROR), "系统错误")
	}

	// 3. 检查是否超过限制
	if count >= constants.SmsCodeMaxRetryCount {
		// 获取最早的重试记录时间
		items, err := l.svcCtx.Redis.ZrangeWithScoresCtx(l.ctx, retryKey, 0, 0)
		if err != nil || len(items) == 0 {
			return errors.Wrapf(ErrSmsCodeRetryLimit, "请等待%d秒后重试", int(constants.SmsCodeRetryWindow.Seconds()))
		}

		// 计算冷却时间
		firstTime := int64(items[0].Score)
		cooldown := firstTime + int64(constants.SmsCodeRetryWindow.Seconds()) - now

		if cooldown > 0 {
			return errors.Wrapf(ErrSmsCodeRetryLimit, "请等待%d秒后重试", cooldown)
		}

		// 如果最早记录已过期，删除整个集合
		_, _ = l.svcCtx.Redis.DelCtx(l.ctx, retryKey)
	}

	return nil
}

// 验证存储的验证码
func (l *VerifySmsCodeLogic) verifyStoredCode(redisKey, inputCode string) (bool, error) {
	// 1. 获取存储的验证码
	storedCode, err := l.svcCtx.Redis.GetCtx(l.ctx, redisKey)
	if err != nil {
		logx.WithContext(l.ctx).Errorf("获取验证码失败: key:%s, err:%v", redisKey, err)
		return false, errors.Wrap(xerrs.NewErrCode(xerrs.SERVER_COMMON_ERROR), "系统错误")
	}

	// 2. 验证码不存在或已过期
	if storedCode == "" {
		return false, nil
	}

	// 3. 验证码比对 (忽略空格和大小写)
	cleanInput := strings.TrimSpace(strings.ToLower(inputCode))
	cleanStored := strings.TrimSpace(strings.ToLower(storedCode))

	return cleanInput == cleanStored, nil
}

// 处理验证成功
func (l *VerifySmsCodeLogic) handleSuccessVerification(redisKey, mobile string, scene int32) error {
	// 使用管道批量操作Redis
	err := l.svcCtx.Redis.PipelinedCtx(l.ctx, func(pipe redis.Pipeliner) error {
		// 1. 删除验证码记录
		pipe.Del(l.ctx, redisKey)

		// 2. 清除重试计数器
		retryKey := fmt.Sprintf("%s:%d:%s", constants.RedisKeySmsRetry, scene, mobile)
		pipe.Del(l.ctx, retryKey)

		// 3. 添加业务标记
		bizKey := fmt.Sprintf("%s:%d:%s", constants.RedisKeySmsVerified, scene, mobile)
		pipe.SetEx(l.ctx, bizKey, "1", time.Duration(10*time.Minute.Seconds()))

		return nil
	})

	if err != nil {
		logx.WithContext(l.ctx).Errorf("验证成功操作Redis失败: mobile:%s, err:%v", mobile, err)
		return err
	}

	// 4. 记录验证成功日志
	logx.WithContext(l.ctx).Infof("验证码验证成功: mobile:%s, scene:%d", mobile, scene)
	return nil
}

// 处理验证失败
func (l *VerifySmsCodeLogic) handleFailedVerification(mobile string, scene int32) {
	// 1. 记录失败尝试
	retryKey := fmt.Sprintf("%s:%d:%s", constants.RedisKeySmsRetry, scene, mobile)
	timestamp := time.Now().Unix()

	err := l.svcCtx.Redis.PipelinedCtx(l.ctx, func(pipe redis.Pipeliner) error {
		// 添加当前时间戳到有序集合
		pipe.ZAdd(l.ctx, retryKey, redis.Z{
			Score:  float64(timestamp),
			Member: timestamp,
		})

		// 设置集合过期时间
		pipe.Expire(l.ctx, retryKey, time.Duration(constants.SmsCodeRetryWindow.Seconds()))

		return nil
	})

	if err != nil {
		logx.WithContext(l.ctx).Errorf("记录失败尝试失败: key:%s, err:%v", retryKey, err)
	}

	// 2. 添加短暂冷却时间
	time.Sleep(constants.SmsCodeCooldown)

	// 3. 记录验证失败日志
	logx.WithContext(l.ctx).Infof("验证码验证失败: mobile:%s, scene:%d", mobile, scene)
}
