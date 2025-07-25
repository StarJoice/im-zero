// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.4.0
// - protoc             v3.19.4
// source: verifycode.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.62.0 or later.
const _ = grpc.SupportPackageIsVersion8

const (
	Verifycode_SendSmsCode_FullMethodName   = "/pb.Verifycode/SendSmsCode"
	Verifycode_VerifySmsCode_FullMethodName = "/pb.Verifycode/VerifySmsCode"
)

// VerifycodeClient is the client API for Verifycode service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// 验证码服务
type VerifycodeClient interface {
	// 发送验证码
	SendSmsCode(ctx context.Context, in *SendSmsCodeReq, opts ...grpc.CallOption) (*SendSmsCodeResp, error)
	// 验证验证码
	VerifySmsCode(ctx context.Context, in *VerifySmsCodeReq, opts ...grpc.CallOption) (*VerifySmsCodeResp, error)
}

type verifycodeClient struct {
	cc grpc.ClientConnInterface
}

func NewVerifycodeClient(cc grpc.ClientConnInterface) VerifycodeClient {
	return &verifycodeClient{cc}
}

func (c *verifycodeClient) SendSmsCode(ctx context.Context, in *SendSmsCodeReq, opts ...grpc.CallOption) (*SendSmsCodeResp, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SendSmsCodeResp)
	err := c.cc.Invoke(ctx, Verifycode_SendSmsCode_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *verifycodeClient) VerifySmsCode(ctx context.Context, in *VerifySmsCodeReq, opts ...grpc.CallOption) (*VerifySmsCodeResp, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(VerifySmsCodeResp)
	err := c.cc.Invoke(ctx, Verifycode_VerifySmsCode_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VerifycodeServer is the server API for Verifycode service.
// All implementations must embed UnimplementedVerifycodeServer
// for forward compatibility
//
// 验证码服务
type VerifycodeServer interface {
	// 发送验证码
	SendSmsCode(context.Context, *SendSmsCodeReq) (*SendSmsCodeResp, error)
	// 验证验证码
	VerifySmsCode(context.Context, *VerifySmsCodeReq) (*VerifySmsCodeResp, error)
	mustEmbedUnimplementedVerifycodeServer()
}

// UnimplementedVerifycodeServer must be embedded to have forward compatible implementations.
type UnimplementedVerifycodeServer struct {
}

func (UnimplementedVerifycodeServer) SendSmsCode(context.Context, *SendSmsCodeReq) (*SendSmsCodeResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendSmsCode not implemented")
}
func (UnimplementedVerifycodeServer) VerifySmsCode(context.Context, *VerifySmsCodeReq) (*VerifySmsCodeResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifySmsCode not implemented")
}
func (UnimplementedVerifycodeServer) mustEmbedUnimplementedVerifycodeServer() {}

// UnsafeVerifycodeServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to VerifycodeServer will
// result in compilation errors.
type UnsafeVerifycodeServer interface {
	mustEmbedUnimplementedVerifycodeServer()
}

func RegisterVerifycodeServer(s grpc.ServiceRegistrar, srv VerifycodeServer) {
	s.RegisterService(&Verifycode_ServiceDesc, srv)
}

func _Verifycode_SendSmsCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendSmsCodeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VerifycodeServer).SendSmsCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Verifycode_SendSmsCode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VerifycodeServer).SendSmsCode(ctx, req.(*SendSmsCodeReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Verifycode_VerifySmsCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerifySmsCodeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VerifycodeServer).VerifySmsCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Verifycode_VerifySmsCode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VerifycodeServer).VerifySmsCode(ctx, req.(*VerifySmsCodeReq))
	}
	return interceptor(ctx, in, info, handler)
}

// Verifycode_ServiceDesc is the grpc.ServiceDesc for Verifycode service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Verifycode_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Verifycode",
	HandlerType: (*VerifycodeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendSmsCode",
			Handler:    _Verifycode_SendSmsCode_Handler,
		},
		{
			MethodName: "VerifySmsCode",
			Handler:    _Verifycode_VerifySmsCode_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "verifycode.proto",
}
