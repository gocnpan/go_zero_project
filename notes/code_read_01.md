# `code`阅读
以`user`服务为例

## API
`api/user.go`为程序入口，主要功能有：
- 从`api/etc/user.yaml`读取配置项（`Mysql & CacheRedis & Auth & UserRpc`)
- 注册`rpc & restfl`服务

`api`子目录：
- `etc`：配置文件
- `internal`：网络服务
    - `config`：配置项-结构体
    - `handler`：路由
      - `routes.go` 路由注册，权限校验采用`JWT`
    - `logic`：逻辑处理
      - `loginlogic.go` 登录时生成`JWT`
    - `svc`：网络环境：返回 RPC 客户端-结构体
    - `types`：数据定义-结构体
    
## model
提供数据的`CRUD`功能

## RPC
`rpc/user.go`为服务启动入口，主要功能有：
- 从`api/etc/user.yaml`读取配置项（`Etcd & Mysql & CacheRedis & Salt`)
- 注册数据服务（`mysql & redis`）、RPC服务

`rpc`子目录
- `etc`：配置
- `internal` ：数据服务
  - `config`：配置项：rpc、mysql、redis、盐
  - `logic`：用户服务逻辑：调用model方法增删改查，并校验返回
  - `server`：服务端方法的封装
  - `svc`：存储服务封装
- `user`：grpc自动生成代码
  - `user.pb.go`：pb数据处理
  - `user_grpc.pb.go`：客户端请求、服务端接口及注册方法
- `userclient`：`rpc`客户端，通过`user/user_grpc.pb.go`方法调用`rpc`服务
