package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"im-zero/app/message/cmd/rpc/internal/svc"
	"im-zero/app/message/cmd/rpc/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type PushMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPushMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PushMessageLogic {
	return &PushMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 内部推送请求结构
type InternalPushReq struct {
	UserId  int64   `json:"userId"`
	Message Message `json:"message"`
}

type Message struct {
	Id             int64  `json:"id"`
	FromUserId     int64  `json:"fromUserId"`
	ToUserId       int64  `json:"toUserId"`
	ConversationId string `json:"conversationId"`
	MessageType    int32  `json:"messageType"`
	Content        string `json:"content"`
	Extra          string `json:"extra"`
	Status         int32  `json:"status"`
	CreateTime     int64  `json:"createTime"`
	UpdateTime     int64  `json:"updateTime"`
}

type InternalPushResp struct {
	Success bool `json:"success"`
}

// 推送消息给用户
func (l *PushMessageLogic) PushMessage(in *message.PushMessageReq) (*message.PushMessageResp, error) {
	// 参数验证
	if in.UserId <= 0 {
		l.Logger.Errorf("Invalid user id: %d", in.UserId)
		return &message.PushMessageResp{Success: false}, nil
	}

	if in.Message == nil {
		l.Logger.Errorf("Message is nil for user: %d", in.UserId)
		return &message.PushMessageResp{Success: false}, nil
	}

	// 通过HTTP调用API服务的内部推送接口
	success := l.callInternalPushAPI(in.UserId, in.Message)

	l.Logger.Infof("消息推送请求: 用户=%d, 消息ID=%d, 结果=%v", in.UserId, in.Message.Id, success)

	return &message.PushMessageResp{Success: success}, nil
}

// 调用API服务的内部推送接口
func (l *PushMessageLogic) callInternalPushAPI(userId int64, msg *message.MessageInfo) bool {
	// API服务的内部推送URL (这里假设API服务在8005端口)
	url := "http://localhost:8005/message/v1/internal/push"

	// 构造请求数据
	reqData := InternalPushReq{
		UserId: userId,
		Message: Message{
			Id:             msg.Id,
			FromUserId:     msg.FromUserId,
			ToUserId:       msg.ToUserId,
			ConversationId: msg.ConversationId,
			MessageType:    msg.MessageType,
			Content:        msg.Content,
			Extra:          msg.Extra,
			Status:         msg.Status,
			CreateTime:     msg.CreateTime,
			UpdateTime:     msg.UpdateTime,
		},
	}

	// 序列化为JSON
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		l.Logger.Errorf("Marshal push request failed: %v", err)
		return false
	}

	// 发送HTTP请求
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		l.Logger.Errorf("Call internal push API failed: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		l.Logger.Errorf("Internal push API returned non-200 status: %d", resp.StatusCode)
		return false
	}

	// 解析响应
	var pushResp InternalPushResp
	if err := json.NewDecoder(resp.Body).Decode(&pushResp); err != nil {
		l.Logger.Errorf("Decode push response failed: %v", err)
		return false
	}

	return pushResp.Success
}
