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
   
## `api/internal/handler`模块
### 文件-[loginhandler.go](../../code/service/user/api/internal/handler/loginhandler.go)
1. 解析请求携带的数据：路径参数、头部、表单、body。
   - 代码：[rest/httpx/requests.go](../rest/httpx/requests.go)

2. 挂载上下文：链路跟踪、请求、RPC
   ```go
       LoginLogic{
           Logger: logx.WithContext(ctx), // 日志 记录请求上下文，便于链路跟踪
           ctx:    ctx, // 请求上下文
           svcCtx: svcCtx, // 服务器上下文，配置 & rpc客户端
       }
   ```

3. 调用`api/internal/logic`相应逻辑处理

4. 将状态码、回复消息写入w：可能成功或失败

### 文件-[logic/loginlogic.go](../../code/service/user/api/internal/logic/loginlogic.go)
1. 调用 `user rpc` 客户端`Login`接口，进行登录，无误则进入下一步，否则返回错误信息
   
2. 账号密码无误，生成 `jwt token` 并返回
