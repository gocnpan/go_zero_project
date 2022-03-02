package internal

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
)

// StartOption defines the method to customize http.Server.
type StartOption func(srv *http.Server)

// StartHttp starts a http server.
func StartHttp(host string, port int, handler http.Handler, opts ...StartOption) error {
	return start(host, port, handler, func(srv *http.Server) error {
		return srv.ListenAndServe() // 开启 http 服务
	}, opts...)
}

// StartHttps starts a https server.
func StartHttps(host string, port int, certFile, keyFile string, handler http.Handler,
	opts ...StartOption) error {
	return start(host, port, handler, func(srv *http.Server) error {
		// certFile and keyFile are set in buildHttpsServer
		return srv.ListenAndServeTLS(certFile, keyFile)
	}, opts...)
}

// start
// handler 结构如下，在 go-zero/rest/router/patrouter.go 中定义
// 		patRouter{
//			trees: make(map[string]*search.Tree),
//		}
func start(host string, port int, handler http.Handler, run func(srv *http.Server) error,
	opts ...StartOption) (err error) {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: handler,
	}
	for _, opt := range opts {
		opt(server)
	}

	waitForCalled := proc.AddWrapUpListener(func() { // 关闭监听器
		if e := server.Shutdown(context.Background()); err != nil {
			logx.Error(e)
		}
	})
	defer func() {
		if err == http.ErrServerClosed {
			waitForCalled()
		}
	}()

	return run(server)
}
