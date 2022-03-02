文件：[zrpc/client.go](../zrpc/client.go)
## 简介
1. 生成 grpc 客户端链接（grpc直连 | etcd服务发现）
2. grpc连接

## 详细
由`service/user/api/user.go`的`ctx := svc.NewServiceContext(c)`方法依次调用：
- `service/user/api/internal/svc/servicecontext.go`---》`zrpc.MustNewClient(c.UserRpc)`
- `go-zero/zrpc/client.go`---》`NewClient` 返回 target为 rpc 直连或 etcd 服务发现的`grpc.ClientConn`结构体

### NewClient 方法
`user/api`的配置项：
```yaml
UserRpc:
  App: userapi
  Token: 6jKNZbEpYGeUMAifz10gOnmoty3TV
  Etcd:
    Hosts:
      - etcd:2379
    Key: user.rpc
```

配置权限校验、阻塞、超时等[内容](../zrpc/client.go)

`target, err := c.BuildTarget()`获取 RPC 链接：[code](../zrpc/client.go)
- 有2种形式：
  1. etcd服务发现：host(etcd service name:port) + key	即：discov://etcd:2379/user.rpc
  2. rpc直连：endpoints

`client, err := internal.NewClient(target, opts...)`新建 rpc 连接
