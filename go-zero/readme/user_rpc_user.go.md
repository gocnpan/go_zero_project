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

## gRPC 注册发现相关
参考：[notes/grpc_intro.md](../../notes/grpc_intro.md)


   
