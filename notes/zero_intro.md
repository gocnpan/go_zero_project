# go-zero上手
参考：[单体服务](https://go-zero.dev/cn/monolithic-service.html)

创建`hello_zero`，命令
- `go mod init hello_zero`
- `goctl api new greet`
- `go mod tidy`

其中命令`goctl api new greet`，生成
```tree
$ tree greet
greet
├── etc
│   └── greet-api.yaml
├── greet.api
├── greet.go
└── internal
    ├── config
    │   └── config.go
    ├── handler
    │   ├── greethandler.go
    │   └── routes.go
    ├── logic
    │   └── greetlogic.go
    ├── svc
    │   └── servicecontext.go
    └── types
        └── types.go
```

编写逻辑，在`logic/greetlogic.go`修改
```protobuf
func (l *GreetLogic) Greet(req types.Request) (*types.Response, error) {
    return &types.Response{
        Message: "Hello go-zero",
    }, nil
}
```

启动服务`go run greet.go -f etc/greet-api.yaml`

访问服务`http://localhost:8888/from/you`