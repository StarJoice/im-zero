package xerrs

var message map[uint32]string

func init() {
	message = make(map[uint32]string)
	message[OK] = "SUCCESS"

	// 全局错误消息
	message[SERVER_COMMON_ERROR] = "服务器开小差啦,稍后再来试一试"
	message[PARAM_ERROR] = "参数错误"
	message[TOKEN_EXPIRE_ERROR] = "token失效，请重新登陆"
	message[TOKEN_GENERATE_ERROR] = "生成token失败"
	message[DB_ERROR] = "数据库繁忙,请稍后再试"
	message[DB_UPDATE_AFFECTED_ZERO_ERROR] = "更新数据影响行数为0"
	message[NOT_FOUND] = "资源不存在"
	message[UNAUTHORIZED] = "未授权"
	message[PERMISSION_DENIED] = "权限不足"
	message[RATE_LIMIT_ERROR] = "操作太频繁，请稍后再试"
	message[NETWORK_ERROR] = "网络错误，请检查网络连接"
	message[TIMEOUT_ERROR] = "请求超时，请重试"
	message[INVALID_REQUEST] = "无效请求"
	message[SERVICE_UNAVAILABLE] = "服务不可用"

	// 用户模块错误
	message[USER_NOT_FOUND] = "用户不存在"
	message[USER_ALREADY_EXISTS] = "用户已存在"
	message[INVALID_PASSWORD] = "密码错误"
	message[INVALID_MOBILE] = "手机号格式不正确"
	message[USER_DISABLED] = "用户已被禁用"
	message[FRIEND_ALREADY_EXISTS] = "好友已存在"
	message[NOT_FRIEND_RELATION] = "非好友关系"
	message[FRIEND_REQUEST_NOT_FOUND] = "好友请求不存在"
	message[FRIEND_REQUEST_EXPIRED] = "好友请求已过期"
	message[FRIEND_REQUEST_ALREADY_HANDLED] = "好友请求已被处理"
	message[CANNOT_ADD_SELF_AS_FRIEND] = "不能添加自己为好友"
	message[USER_BLOCKED] = "用户已被拉黑"

	// SMS模块错误
	message[VERIFY_CODE_ERROR] = "验证码错误"
	message[VERIFY_CODE_EXPIRED] = "验证码已过期"
	message[VERIFY_CODE_RETRY_LIMIT] = "验证码重试次数过多"
	message[SMS_SEND_FAILED] = "短信发送失败"
	message[SMS_TEMPLATE_NOT_FOUND] = "短信模板不存在"
	message[SMS_RATE_LIMIT] = "短信发送频率限制"

	// 群组模块错误
	message[GROUP_NOT_FOUND] = "群组不存在"
	message[GROUP_ALREADY_DISSOLVED] = "群组已解散"
	message[GROUP_MEMBER_FULL] = "群成员已满"
	message[GROUP_MEMBER_NOT_FOUND] = "用户不在群组中"
	message[GROUP_MEMBER_ALREADY_EXISTS] = "用户已在群组中"
	message[GROUP_OWNER_CANNOT_LEAVE] = "群主不能退出群组"
	message[GROUP_PERMISSION_DENIED] = "群组权限不足"
	message[GROUP_MEMBER_MUTED] = "用户已被禁言"
	message[GROUP_INVITE_NOT_ALLOWED] = "不允许邀请用户"
	message[GROUP_JOIN_APPROVAL_REQUIRED] = "入群需要审批"
	message[GROUP_NAME_TOO_LONG] = "群名太长"
	message[GROUP_DESCRIPTION_TOO_LONG] = "群描述太长"
	message[GROUP_ALREADY_EXISTS] = "群组已存在"

	// 消息模块错误
	message[MESSAGE_NOT_FOUND] = "消息不存在"
	message[MESSAGE_SEND_FAILED] = "消息发送失败"
	message[MESSAGE_RECALL_TIMEOUT] = "消息撤回超时"
	message[MESSAGE_ALREADY_RECALLED] = "消息已撤回"
	message[CONVERSATION_NOT_FOUND] = "会话不存在"
	message[MESSAGE_CONTENT_TOO_LONG] = "消息内容太长"
	message[INVALID_MESSAGE_TYPE] = "无效消息类型"
	message[MESSAGE_DELETE_FAILED] = "消息删除失败"
}

func MapErrMsg(errcode uint32) string {
	if msg, ok := message[errcode]; ok {
		return msg
	} else {
		return "服务器开小差啦,稍后再来试一试"
	}
}

func IsCodeErr(errcode uint32) bool {
	if _, ok := message[errcode]; ok {
		return true
	} else {
		return false
	}
}
