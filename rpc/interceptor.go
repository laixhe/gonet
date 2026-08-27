package rpc

import (
	"context"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryUnaryServerInterceptor 捕获 handler panic，转成 codes.Internal 错误返回，
// 避免 panic 直接打崩服务进程。用法：
//
//	NewServer(addr, cert, key, WithServerOption(grpc.ChainUnaryInterceptor(
//		rpc.RecoveryUnaryServerInterceptor(),
//		rpc.LoggingUnaryServerInterceptor(),
//	)))
func RecoveryUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				logf("panic in %s: %v\n%s", info.FullMethod, r, debug.Stack())
				err = status.Errorf(codes.Internal, "panic: %v", r)
			}
		}()
		return handler(ctx, req)
	}
}

// LoggingUnaryServerInterceptor 打印每个 RPC 的方法、结果码与耗时。
func LoggingUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		logf("grpc %s code=%s dur=%s", info.FullMethod, status.Code(err), time.Since(start))
		return resp, err
	}
}

// RecoveryStreamServerInterceptor 捕获流式 handler panic，转成 codes.Internal 错误返回，
// 避免 stream 中的 panic 直接打崩服务进程。用法与一元版对称：
//
//	NewServer(addr, cert, key, WithServerOption(grpc.ChainStreamInterceptor(
//		rpc.RecoveryStreamServerInterceptor(),
//		rpc.LoggingStreamServerInterceptor(),
//	)))
func RecoveryStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logf("panic in stream %s: %v\n%s", info.FullMethod, r, debug.Stack())
				err = status.Errorf(codes.Internal, "panic: %v", r)
			}
		}()
		return handler(srv, ss)
	}
}

// LoggingStreamServerInterceptor 打印每个流式 RPC 的方法、结果码与耗时。
func LoggingStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		start := time.Now()
		err = handler(srv, ss)
		logf("grpc stream %s code=%s dur=%s", info.FullMethod, status.Code(err), time.Since(start))
		return err
	}
}

// LoggingUnaryClientInterceptor 打印每个客户端一元 RPC 的方法、结果码与耗时。
// 与 LoggingUnaryServerInterceptor 对称，便于在客户端侧观测调用。
// 用法：
//
//	NewClient(addr, WithClientOption(grpc.WithUnaryInterceptor(rpc.LoggingUnaryClientInterceptor())))
func LoggingUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		logf("grpc client %s code=%s dur=%s", method, status.Code(err), time.Since(start))
		return err
	}
}
