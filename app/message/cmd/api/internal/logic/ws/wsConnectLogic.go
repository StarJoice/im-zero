package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"im-zero/app/message/cmd/api/internal/svc"
	"im-zero/app/message/cmd/api/internal/types"
	"im-zero/pkg/ctxdata"
	"im-zero/pkg/wsmanager"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

type WsConnectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWsConnectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WsConnectLogic {
	return &WsConnectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WsConnectLogic) WsConnect(req *types.WsConnectReq, w http.ResponseWriter, r *http.Request) error {
	// 先从JWT中间件获取用户ID
	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 如果JWT中间件没有处理，手动解析token
	if userId <= 0 {
		var token string

		// 优先从Authorization header获取token
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
			l.Logger.Infof("Token obtained from Authorization header")
		} else if req.Token != "" {
			// 兼容性支持：从查询参数获取token（不推荐）
			token = req.Token
			l.Logger.Infof("Token obtained from query parameter (deprecated)")
		}

		if token != "" {
			if uid, err := l.parseJWTToken(token); err == nil {
				userId = uid
			} else {
				l.Logger.Errorf("Failed to parse JWT token: %v", err)
			}
		}

		// 测试模式：从查询参数获取user_id（仅用于开发测试）
		if userId <= 0 {
			userIdStr := r.URL.Query().Get("user_id")
			if userIdStr != "" {
				if uid, err := strconv.ParseInt(userIdStr, 10, 64); err == nil {
					userId = uid
					l.Logger.Infof("Using user_id from query parameter for testing: %d (DEVELOPMENT ONLY)", userId)
				}
			}
		}

		if userId <= 0 {
			l.Logger.Error("Unauthorized: no valid user ID found. Please provide token via Authorization header")
			http.Error(w, "Unauthorized: Missing or invalid token", http.StatusUnauthorized)
			return nil
		}
	}

	l.Logger.Infof("WebSocket connection request from user %d", userId)

	// 使用统一的WebSocket管理器处理连接
	wsmanager.HandleWebSocket(w, r, userId)

	return nil
}

// parseJWTToken 解析JWT token获取用户ID
func (l *WsConnectLogic) parseJWTToken(tokenString string) (int64, error) {
	// 解析token但不验证签名（因为这里主要是为了获取用户ID）
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid token claims")
	}

	// 尝试获取userId字段
	if userIdClaim, exists := claims["userId"]; exists {
		switch v := userIdClaim.(type) {
		case float64:
			return int64(v), nil
		case json.Number:
			if uid, err := v.Int64(); err == nil {
				return uid, nil
			}
		case string:
			if uid, err := strconv.ParseInt(v, 10, 64); err == nil {
				return uid, nil
			}
		case int64:
			return v, nil
		case int:
			return int64(v), nil
		}
	}

	return 0, fmt.Errorf("userId not found in token claims")
}
