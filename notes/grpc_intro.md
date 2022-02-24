# GRPC使用
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