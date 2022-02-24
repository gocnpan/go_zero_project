// main.go
package main

import (
	"google.golang.org/grpc"
	proto2 "hello_grpc/proto"
	"hello_grpc/server/controller"
	"log"
	"net"
)

const Address = "0.0.0.0:9090"

func main() {
	listen, err := net.Listen("tcp", Address)
	if err != nil {
		log.Fatal("Failed to listen: %v", err)
	}
	s := grpc.NewServer() // 启动新 rpc 服务
	//服务注册
	proto2.RegisterHelloServer(s, &controller.HelloController{}) // 注册服务端方法
	log.Println("Listen on " + Address)
	err = s.Serve(listen)
	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}