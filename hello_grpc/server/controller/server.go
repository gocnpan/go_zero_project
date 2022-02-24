// hello_server.go
package controller

import (
	"context"
	"fmt"
	proto2 "hello_grpc/proto"
)

// HelloController 是对 hello.pb.go 文件中 接口 HelloServer 的实现
type HelloController struct {

}

func (h *HelloController) SayHello(ctx context.Context, in  *proto2.HelloRequest) (*proto2.HelloResponse, error){
	return &proto2.HelloResponse{Message : fmt.Sprintf("%s", in.Name)}, nil
}

func (h *HelloController) LotsOfReplies(in *proto2.HelloRequest, stream proto2.Hello_LotsOfRepliesServer)  error{
	for i := 0; i < 10; i++ {
		stream.Send(&proto2.HelloResponse{
			Message: fmt.Sprintf("%s %s %d", in.Name, "Reply", i),
		})
	}
	return nil
}