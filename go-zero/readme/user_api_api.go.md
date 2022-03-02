# 文件
以`user/api/user.go`为切入口，阅读源代码
[`code/service/user/api/user.go`](../../code/service/user/api/user.go)

## `main`方法
1. 从运行命令中获取配置文件信息，配置文件路径默认为`etc/user.yaml`

2. 从文件中读取并填入配置项：`conf.MustLoad(*configFile, &c)` \
详见：[core_conf](core_conf.md) \
   
3. 初始化`gRPC`：`ctx := svc.NewServiceContext(c)` \
详见：[zrpc_client](zrpc_client.md)
   
4. 初始化`restful`：`server := rest.MustNewServer(c.RestConf)` \
详见：[rest_server](rest_server.md)
   
5. `restful`路由注册：按路由配置方法的类别，分别注册 \
   - 代码：`handler.RegisterHandlers(server, ctx)` \
   - `RegisterHandlers`方法中`rest.WithJwt(serverCtx.Config.Auth.AccessSecret)`为路由配置jwt参数，附加相关信息
   - 详见：[rest_server](rest_server.md)
   
