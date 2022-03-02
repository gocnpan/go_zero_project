package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf  // restful 结构体
	Auth struct { // 权限校验 结构体
		AccessSecret string // 密钥
		AccessExpire int64 // 有效期
	}
	UserRpc zrpc.RpcClientConf
}
