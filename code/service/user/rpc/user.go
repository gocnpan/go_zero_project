package main

import (
	"flag"
	"fmt"

	"mall/service/user/rpc/internal/config"
	"mall/service/user/rpc/internal/server"
	"mall/service/user/rpc/internal/svc"
	"mall/service/user/rpc/user"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/user.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	// 注册数据服务
	// 创建 mysql 连接 struct 对象
	// 传入 mysql & cache redis 连接，注册 user model
	ctx := svc.NewServiceContext(c)
	srv := server.NewUserServer(ctx) // 获取服务端RPC 方法

	// 校验 redis 配置
	// 获取运行环境指标对象 metrics，用于监测、降载等
	// 有 etcd 时，在 etcd 注册 grpc 服务（并定时更新）
	// 配置 rpc 服务
	// 配置 rpc 服务拦截器：自适应负载均衡、超时、权限拦截
	// 启动：日志、调测环境配置、链路跟踪、远程监控
	// 	rpcServer := &RpcServer{
	//		server:   server,  // rpc 服务对象
	//		register: register, // rpc 方法注册 func
	//	}
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		user.RegisterUserServer(grpcServer, srv) // 在grpc中注册 方法

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer) // 服务映射
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start() // 调用 MustNewServer 传入的 internal.RegisterFn
}
