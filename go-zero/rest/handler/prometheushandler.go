package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/metric"
	"github.com/zeromicro/go-zero/core/prometheus"
	"github.com/zeromicro/go-zero/core/timex"
	"github.com/zeromicro/go-zero/rest/internal/response"
)

const serverNamespace = "http_server"

var (
	metricServerReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "http server requests duration(ms).",
		Labels:    []string{"path"}, // 监控指标
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000}, // 直方图分布中，统计的桶
	})

	metricServerReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "http server requests error count.",
		Labels:    []string{"path", "code"},  // 监控指标：直接在记录指标 incr() 即可
	})
)

// PrometheusHandler returns a middleware that reports stats to prometheus.
func PrometheusHandler(path string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if !prometheus.Enabled() {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 请求进入的时间
			startTime := timex.Now()
			cw := &response.WithCodeResponseWriter{Writer: w}
			defer func() {
				// 请求返回的时间
				metricServerReqDur.Observe(int64(timex.Since(startTime)/time.Millisecond), path)
				metricServerReqCodeTotal.Inc(path, strconv.Itoa(cw.Code))
			}()
			// 中间件放行，执行完后续中间件和业务逻辑。重新回到这，做一个完整请求的指标上报
			// [🧅：洋葱模型]
			next.ServeHTTP(cw, r)
		})
	}
}
