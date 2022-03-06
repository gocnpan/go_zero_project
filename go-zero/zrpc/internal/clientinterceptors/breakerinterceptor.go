package clientinterceptors

import (
	"context"
	"path"

	"github.com/zeromicro/go-zero/core/breaker"
	"github.com/zeromicro/go-zero/zrpc/internal/codes"
	"google.golang.org/grpc"
)

// BreakerInterceptor is an interceptor that acts as a circuit breaker.
func BreakerInterceptor(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// target + 方法名
	// 基于请求方法进行熔断
	breakerName := path.Join(cc.Target(), method)
	return breaker.DoWithAcceptable(breakerName, func() error {
		// 真正执行调用
		return invoker(ctx, method, req, reply, cc, opts...)
		// codes.Acceptable判断哪种错误需要加入熔断错误计数
	}, codes.Acceptable)
}
