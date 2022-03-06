package serverinterceptors

import (
	"context"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/metric"
	"github.com/zeromicro/go-zero/core/prometheus"
	"github.com/zeromicro/go-zero/core/timex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const serverNamespace = "rpc_server"

var (
	metricServerReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "rpc server requests duration(ms).",
		Labels:    []string{"method"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})

	metricServerReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "rpc server requests code count.",
		Labels:    []string{"method", "code"},
	})
)

// UnaryPrometheusInterceptor reports the statistics to the prometheus server.
// 拦截器主要是对服务的监控指标进行收集
// 这里主要是对RPC方法的耗时和调用错误进行收集
// 这里主要使用了 Prometheus 的 Histogram 和 Counter 数据类型
func UnaryPrometheusInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if !prometheus.Enabled() {
		return handler(ctx, req)
	}

	startTime := timex.Now()
	resp, err := handler(ctx, req)
	metricServerReqDur.Observe(int64(timex.Since(startTime)/time.Millisecond), info.FullMethod)
	metricServerReqCodeTotal.Inc(info.FullMethod, strconv.Itoa(int(status.Code(err))))
	return resp, err
}
