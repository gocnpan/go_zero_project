package main

import (
	"flag"
	"fmt"

	"mall/service/user/api/internal/config"
	"mall/service/user/api/internal/handler"
	"mall/service/user/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

// 从 etc 配置文件中读取 config
var configFile = flag.String("f", "etc/user.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)  // 将配置信息 赋予 c

	ctx := svc.NewServiceContext(c) // rpc 环境
	server := rest.MustNewServer(c.RestConf) // restful 环境
	defer server.Stop()

	handler.RegisterHandlers(server, ctx) // 注册 api - 路由注册

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
