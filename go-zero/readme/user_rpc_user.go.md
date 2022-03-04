# 文件
以`user/rpc/user.go`为切入口，阅读源代码
[`code/service/user/rpc/user.go`](../../code/service/user/rpc/user.go)

## `main`方法
1. 读取并设置配置项

2. 获取数据服务上下文
    ```go
        conn := sqlx.NewMysql(c.Mysql.DataSource) // mysql 连接     
        ServiceContext{
            Config:    c,
              // 注册用户 model（含 持久化-msql & 缓存-redis）
            UserModel: model.NewUserModel(conn, c.CacheRedis),
        }
    ```
   详见：[core_stores.md](core_stores.md)
   
3. 获取`grpc`服务上下文
   ```go
   type UserServer struct {
       svcCtx *svc.ServiceContext
       user.UnimplementedUserServer
   }
   ```
   - 实现：Login、Register、UserInfo服务
   
4. 配置`grpc`服务，传入`grpc`方法注册func
   - 校验 redis 配置
   - 获取运行环境指标对象 metrics，用于监测、降载等
   - 有 etcd 时，在 etcd 注册 grpc 服务（并定时更新）
   - 配置 rpc 服务
   - 配置 rpc 服务拦截器：[自适应负载均衡](https://www.jianshu.com/p/71a3569ed205) 、超时、权限拦截
   - 启动：日志、调测环境配置、链路跟踪、远程监控
   ```go
   rpcServer := &RpcServer{
       server:   server, // rpc 服务对象
       register: register, // rpc 服务方法注册 func 
   }
   ```

## gRPC 注册发现相关
参考：[notes/grpc_intro.md](../../notes/grpc_intro.md)


   
