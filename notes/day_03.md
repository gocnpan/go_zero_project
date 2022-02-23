# 用户服务

## 环境要求
1. 安装go-zero & goctl
   - `go install github.com/zeromicro/go-zero/tools/goctl@latest`
   - 注意，此处安装，是安装到`$GOPATH/bin`目录下
   - 因此在win10，需要将`$GOPATH/bin`添加到环境变量中

2. 安装 protoc
   - 去 [这里](https://github.com/google/protobuf/releases) 下载对应的protoc，我这里下的是protoc-3.19.0-win64.zip
   - 下好之后解压就行，然后把bin里面的protoc.exe加入到环境变量，并且把protoc.exe拷贝到C:\Windows\System32
 
3. 安装 protoc-gen-go（使用）
   - go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
   - go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
   
## 容器使用：[Docker教程](https://www.runoob.com/docker/docker-container-usage.html)
1. 列出容器：`docker ps` 

2. 使用容器（在gonivinck项目的容器）：`docker exec -it 243c32535da7 /bin/bash`，其中`243c32535da7`是目标容器id

3. 宿主目录挂在到容器：使用绝对路径如`/e/code/golang/go_zero_project/code`
   - 等价于`E:\code\golang\go_zero_project\code`

4. 在golang容器中运行(通过步骤2进入golang容器)：
   - `go run ./service/user/rpc/user.go -f ./service/user/rpc/etc/user.yaml`
   - `go run ./service/user/api/user.go -f ./service/user/api/etc/user.yaml`
   - 运行上述2个命令，有先后顺序
