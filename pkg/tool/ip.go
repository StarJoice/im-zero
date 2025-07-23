package tool

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const ClientIPKey = "client_ip"

// ClientIPInterceptor 是一个 gRPC 服务器拦截器，用于从元数据中提取客户端 IP
func ClientIPInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 从元数据中获取客户端 IP
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if clientIPs, exists := md["client_ip"]; exists && len(clientIPs) > 0 {
			ctx = context.WithValue(ctx, ClientIPKey, clientIPs[0])
		}
	}

	return handler(ctx, req)
}
