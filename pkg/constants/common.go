package constants

import "time"

const (
	// 验证码相关常量
	SmsCodeExpireTime    = 5 * time.Minute  // 验证码过期时间
	SmsCodeMaxRetryCount = 5                // 最大重试次数
	SmsCodeRetryWindow   = 5 * time.Minute  // 重试时间窗口
	SmsCodeCooldown      = 30 * time.Second // 验证失败冷却时间

	// 用户认证类型
	UserAuthTypeSystem  = "system"
	UserAuthTypeWechat  = "wechat"
	UserAuthTypeWeibo   = "weibo"
	UserAuthTypeQQ      = "qq"

	// JWT Token相关
	JwtTokenExpire     = 7 * 24 * time.Hour // Token过期时间 7天
	JwtRefreshDuration = 1 * time.Hour      // 刷新时间 1小时

	// Redis Key前缀
	RedisKeySmsCode     = "sms:code"
	RedisKeySmsRetry    = "sms:verify:retry"
	RedisKeySmsVerified = "sms:verified"
	RedisKeyUserToken   = "user:token"

	// 短信模板场景
	SmsSceneRegister      = 1 // 注册
	SmsSceneLogin         = 2 // 登录
	SmsSceneResetPassword = 3 // 重置密码
	SmsSceneBindPhone     = 4 // 绑定手机号

	// 短信模板ID
	SmsTemplateRegister      = "SMS_REGISTER_VERIFY"
	SmsTemplateLogin         = "SMS_LOGIN_VERIFY"
	SmsTemplateResetPassword = "SMS_RESET_PASSWORD"
	SmsTemplateBindPhone     = "SMS_BIND_PHONE"
	SmsTemplateDefault       = "SMS_DEFAULT_VERIFY"
)

// 获取短信模板ID
func GetSmsTemplateByScene(scene int32) string {
	switch scene {
	case SmsSceneRegister:
		return SmsTemplateRegister
	case SmsSceneLogin:
		return SmsTemplateLogin
	case SmsSceneResetPassword:
		return SmsTemplateResetPassword
	case SmsSceneBindPhone:
		return SmsTemplateBindPhone
	default:
		return SmsTemplateDefault
	}
}