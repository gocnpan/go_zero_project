# GRPC使用 & go-zero的应用
参考：https://www.jianshu.com/p/5745994e6707

## 下载安装
1. 安装protoc，参考第三天的[内容](day_03.md)
2. 安装Go protobuf插件
   - go get -u github.com/golang/protobuf/proto
   - go get -u github.com/golang/protobuf/protoc-gen-go
   - go get -u google.golang.org/protobuf
   - go get -u google.golang.org/grpc
   
3. 编写hello.proto文件
   - 生成的文件中，包含接口、数据结构、未实现方法时错误提示等内容
```protobuf
//指定proto版本
syntax = "proto3";
//指定包名
package hello;
option go_package="./hello";

//定义Hello服务
service Hello{
  //定义SayHello方法
  rpc SayHello(HelloRequest) returns (HelloResponse){}
  //定义LotsOfReplies方法
  rpc LotsOfReplies(HelloRequest) returns (stream HelloResponse){}
}

//HelloRequest请求结构
message HelloRequest{
  string name = 1;
}
//HelloResponse响应结构
message HelloResponse{
  string message = 1;
}
```
4. 生成代码（需要cd到文件`hello.proto`所在目录）：`protoc -I . --go_out=plugins=grpc:. ./hello.proto`
5. 编写服务端提供接口的代码
   - 实现服务端方法（对接口的实现）
   - 注册并启动RPC
   
6. 编写客户端请求接口的代码
   - 注册并启动客户端
   
## go-zero应用
`gRPC`，服务端通过`etcd`注册服务，客户端也通过`etcd`发现服务，实现负载均衡。
- `grpc` 服务端，当配置项`Auth`为真时，从redis获取校验信息，填写密钥相关信息，参考：[day_07.md](day_07.md)
   ```yaml
   Auth: true               # 是否开启 Auth 验证
   StrictControl: true      # 是否开启严格模式
   Redis:                   # 指定 Redis 服务
     Key: rpc:auth:user     # 指定 Key 应为 hash 类型
     Host: redis:6379
     Type: node
     Pass:
   ```
  
- `grpc`服务端，通过`etcd`注册服务，以便让客户端发现服务
   ```yaml
   Etcd:
     Hosts:
       - etcd:2379
     Key: user.rpc
   ```
  
- `grpc`客户端，通过`etcd`发现服务，涉及权限认证，参考：[day_07.md](day_07.md)
   ```yaml
   UserRpc:
     App: userapi  # 在需要 Auth 认证的时候配置
     Token: 6jKNZbEpYGeUMAifz10gOnmoty3TV # 在需要 Auth 认证的时候配置
     Etcd:
       Hosts:
         - etcd:2379
       Key: user.rpc
   ```
  
## gRPC实现负载均衡[原理](https://www.jianshu.com/p/17a9373546a4)
实现基于版本（version）的grpc负载均衡器，了解过程后可自己实现更多的负载均衡功能
- 注册中心
    - Etcd Lease 是一种检测客户端存活状况的机制。 群集授予具有生存时间的租约。 如果etcd 群集在给定的TTL 时间内未收到keepAlive，则租约到期。 为了将租约绑定到键值存储中，每个key 最多可以附加一个租约
- 服务注册 (注册服务)
    - 定时把本地服务（APP）地址,版本等信息注册到服务器
- 服务发现 (客户端发起服务解析请求（APP）)
    - 查询注册中心（APP）下有那些服务
    - 并向所有的服务建立HTTP2长链接
    - 通过Etcd watch 监听服务（APP），通过变化更新链接
- 负载均衡 (客户端发起请求（APP）)
    - 负载均衡选择合适的服务（APP HTTP2长链接）
    - 发起调用