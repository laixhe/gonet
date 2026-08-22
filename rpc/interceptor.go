package rpc

import (
	"context"
	"log"
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
				log.Printf("panic in %s: %v\n%s", info.FullMethod, r, debug.Stack())
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
		log.Printf("grpc %s code=%s dur=%s", info.FullMethod, status.Code(err), time.Since(start))
		return resp, err
	}
}
