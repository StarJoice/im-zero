package logic

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"im-zero/app/verifycode/cmd/rpc/internal/svc"
	"im-zero/app/verifycode/cmd/rpc/verifycode"
	"im-zero/pkg/constants"
	"im-zero/pkg/sms"
	"im-zero/pkg/tool"
)

type SendSmsCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendSmsCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendSmsCodeLogic {
	return &SendSmsCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}
func (l *SendSmsCodeLogic) SendSmsCode(in *verifycode.SendSmsCodeReq) (*verifycode.SendSmsCodeResp, error) {
	// 1. 验证输入 （ 手机号是否符合格式 和 场景类型是否符合要求）
	res := tool.ValidateMobile(in.Mobile)
	if res == false {
		return nil, status.Error(codes.InvalidArgument, "输入手机号格式不正确")
	}
	if in.Scene <= 0 {
		return nil, status.Error(codes.InvalidArgument, "场景类型无效")
	}

	// 2. 生成验证码
	code := tool.Krand(6, tool.KC_RAND_KIND_NUM)

	// 3. 生成唯一key
	codeKey, _ := uuid.GenerateUUID()

	// 4. 存储验证码到Redis
	redisKey := fmt.Sprintf("%s:%d:%s:%s", constants.RedisKeySmsCode, in.Scene, in.Mobile, codeKey)
	err := l.svcCtx.Redis.SetexCtx(l.ctx, redisKey, code, int(constants.SmsCodeExpireTime.Seconds()))
	if err != nil {
		logx.WithContext(l.ctx).Errorf("存储验证码失败: mobile:%s, err:%v", in.Mobile, err)
		return nil, status.Error(codes.Internal, "系统错误")
	}

	// 5. 获取短信模板
	templateID := constants.GetSmsTemplateByScene(in.Scene)
	tpl, exists := l.svcCtx.SmsTemplates.GetTemplate(templateID)
	if !exists || !tpl.Enabled {
		logx.WithContext(l.ctx).Errorf("短信模板不可用: scene:%d, template:%s", in.Scene, templateID)
		return nil, status.Error(codes.FailedPrecondition, "短信服务不可用")
	}

	// 6. 准备模板参数
	params := map[string]string{
		"code": code,
	}

	// 7. 验证模板参数
	if err := l.svcCtx.SmsTemplates.ValidateParams(templateID, params); err != nil {
		logx.WithContext(l.ctx).Errorf("模板参数验证失败: template:%s, params:%v", templateID, params)
		return nil, status.Error(codes.Internal, "系统配置错误")
	}

	// 8. 发送短信
	err = l.svcCtx.SmsService.Send(l.ctx, in.Mobile, templateID, params)
	if err != nil {
		logx.WithContext(l.ctx).Errorf("短信发送失败: mobile:%s, err:%v", in.Mobile, err)

		// 处理特定错误
		var smsErr *sms.SmsError
		if errors.As(err, &smsErr) {
			switch smsErr.Code {
			case sms.ErrCodeRateLimit, sms.ErrCodeIpRateLimit:
				return nil, status.Error(codes.ResourceExhausted, smsErr.Message)
			case sms.ErrCodeBlacklisted:
				return nil, status.Error(codes.PermissionDenied, smsErr.Message)
			default:
				return nil, status.Error(codes.Internal, "短信发送失败")
			}
		}

		return nil, status.Error(codes.Internal, "短信发送失败")
	}

	logx.WithContext(l.ctx).Infof("短信验证码已发送: mobile:%s, scene:%d", in.Mobile, in.Scene)

	// 9. 返回验证码key
	return &verifycode.SendSmsCodeResp{
		CodeKey: codeKey,
	}, nil
}
