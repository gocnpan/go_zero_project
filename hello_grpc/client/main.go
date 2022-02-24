package main

import (
	"context"
	"google.golang.org/grpc"
	proto2 "hello_grpc/proto"
	"io"
	"log"
)

const Address = "0.0.0.0:9090"

func main() {
	conn, err := grpc.Dial(Address, grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	// 一定执行
	defer conn.Close() //
	//初始化客户端
	c := proto2.NewHelloClient(conn)
	//调用SayHello方法
	res, err := c.SayHello(context.Background(), &proto2.HelloRequest{Name: "hello GRPC"})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(res.Message)
	//调用LotsOfReplies方法
	stream, err := c.LotsOfReplies(context.Background(), &proto2.HelloRequest{Name: "Hello GRPC"})
	if err != nil {
		log.Fatalln(err)
	}
	for true {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("stream.Recv: %v", err)
		}
		log.Printf("%s", res.Message)
	}
}
