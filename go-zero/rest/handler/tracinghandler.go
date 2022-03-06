package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// TracingHandler return a middleware that process the opentelemetry.
// 将 header -> carrier，获取 header 中的traceId等信息
// 开启一个新的 span，并把「traceId，spanId」封装在context中
// 从上述的 carrier「也就是header」获取traceId，spanId。
// 看header中是否设置
// 如果没有设置，则随机生成返回
// 从 request 中产生新的ctx，并将相应的信息封装在 ctx 中，返回
// 从上述的 context，拷贝一份到当前的 request
func TracingHandler(serviceName, path string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		propagator := otel.GetTextMapPropagator()
		tracer := otel.GetTracerProvider().Tracer(trace.TraceName)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			spanName := path
			if len(spanName) == 0 {
				spanName = r.URL.Path
			}
			spanCtx, span := tracer.Start(
				ctx,
				spanName,
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
				oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(
					serviceName, spanName, r)...),
			)
			defer span.End()

			// convenient for tracking error messages
			sc := span.SpanContext()
			if sc.HasTraceID() {
				w.Header().Set(trace.TraceIdKey, sc.TraceID().String())
			}

			next.ServeHTTP(w, r.WithContext(spanCtx))
		})
	}
}
