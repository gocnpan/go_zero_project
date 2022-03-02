package rest

import (
	"net/http"
	"time"
)

type (
	// Middleware defines the middleware method.
	Middleware func(next http.HandlerFunc) http.HandlerFunc

	// A Route is a http route.
	Route struct {
		Method  string
		Path    string
		Handler http.HandlerFunc
	}

	// RouteOption defines the method to customize a featured route.
	RouteOption func(r *featuredRoutes)

	jwtSetting struct {
		enabled    bool
		secret     string
		prevSecret string // 旧密钥 用于新旧密钥切换的过渡期
	}

	signatureSetting struct {
		SignatureConf
		enabled bool
	}

	featuredRoutes struct {
		timeout   time.Duration // 超时
		priority  bool // 优先权
		jwt       jwtSetting
		signature signatureSetting // 签名
		routes    []Route // 路由
	}
)
