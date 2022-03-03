package handler

import (
	"net/http"

	"mall/service/user/api/internal/logic"
	"mall/service/user/api/internal/svc"
	"mall/service/user/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginRequest
		// 解析请求 r 携带的数据，给req：路径参数、头部、表单、body
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		// NewLoginLogic 挂载上下文：链路跟踪、请求、RPC
		// 	return LoginLogic{
		//		Logger: logx.WithContext(ctx), // 日志 记录请求上下文，便于链路跟踪
		//		ctx:    ctx, // 请求上下文
		//		svcCtx: svcCtx, // 服务器上下文，配置 & rpc客户端
		//	}
		l := logic.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(req)
		if err != nil {
			//
			httpx.Error(w, err)
		} else {
			// 将成功状态码、回复消息写入w
			httpx.OkJson(w, resp)
		}
	}
}
