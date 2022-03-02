package service

import (
	"log"

	"github.com/zeromicro/go-zero/core/load"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/prometheus"
	"github.com/zeromicro/go-zero/core/stat"
	"github.com/zeromicro/go-zero/core/trace"
)

const (
	// DevMode means development mode.
	DevMode = "dev"
	// TestMode means test mode.
	TestMode = "test"
	// RtMode means regression test mode.
	RtMode = "rt"
	// PreMode means pre-release mode.
	PreMode = "pre"
	// ProMode means production mode.
	ProMode = "pro"
)

// A ServiceConf is a service config.
type ServiceConf struct {
	Name       string
	Log        logx.LogConf
	Mode       string            `json:",default=pro,options=dev|test|rt|pre|pro"`
	MetricsUrl string            `json:",optional"`
	Prometheus prometheus.Config `json:",optional"`
	Telemetry  trace.Config      `json:",optional"`
}

// MustSetUp sets up the service, exits on error.
func (sc ServiceConf) MustSetUp() {
	if err := sc.SetUp(); err != nil {
		log.Fatal(err)
	}
}

// SetUp sets up the service.
func (sc ServiceConf) SetUp() error {
	// 配置 日志 服务名
	if len(sc.Log.ServiceName) == 0 {
		sc.Log.ServiceName = sc.Name
	}
	// 启动 日志
	if err := logx.SetUp(sc.Log); err != nil {
		return err
	}

	sc.initMode() // 根据运行环境配置调测环境
	// 配置 prometheus
	// 如果没有 host 则不启用
	// 一次性调用
	prometheus.StartAgent(sc.Prometheus)

	// openTelemetry 配置：链路追踪 zipkin | jaeger
	if len(sc.Telemetry.Name) == 0 {
		sc.Telemetry.Name = sc.Name
	}
	trace.StartAgent(sc.Telemetry)

	// 远程监控 / 报告？
	if len(sc.MetricsUrl) > 0 {
		stat.SetReportWriter(stat.NewRemoteWriter(sc.MetricsUrl))
	}

	return nil
}

func (sc ServiceConf) initMode() {
	switch sc.Mode {
	// 模式：开发、测试、回退、灰度
	case DevMode, TestMode, RtMode, PreMode:
		load.Disable() // 禁止自适应降载
		stat.SetReporter(nil)
	}
}
