# 功能
根据传入的文件，将配置项填入配置参数（v）

## config.go `MustLoad` --》 `LoadConfig`
以 `user` 启动 `user api` 服务为例，在`user/api`目录，输入命令：`go run user.go -f etc/user.yaml`
其中，`-f` flag 为配置 `config` 的文件路径。

`user.go`调用`conf.MustLoad(*configFile, &c) `方法，传入配置文件路径、配置项指针

配置项内容如下
```go
type Config struct {
	rest.RestConf  // restful 结构体
	Auth struct { // 权限校验结构体
		AccessSecret string
		AccessExpire int64
	}
	UserRpc zrpc.RpcClientConf // rpc 结构体
}
```

根据不同的文件类型，调用不同的方法，通过`hash法`分配不同方法
```go
var loaders = map[string]func([]byte, interface{}) error{
	".json": LoadConfigFromJsonBytes,
	".yaml": LoadConfigFromYamlBytes,
	".yml":  LoadConfigFromYamlBytes,
}
```
更加直观的写法：
```go
var loaders map[string]func([]byte, interface{}) error
loaders = {".json": LoadConfigFromJsonBytes ...}
```

此次以`yaml`为例

## `mapping/yamlunmarshaler.go`
将`yaml`解析为map再转换为struct

关键代码在`go-zero/core/mapping/unmarshaler.go`中的`unmarshalWithFullName`方法

根据配置项结构，一一填充相应的内容。

部分配置项设置了默认值或选填值，如`RestConf`
```go
	RestConf struct {
		service.ServiceConf
		Host     string `json:",default=0.0.0.0"`
		Port     int
		CertFile string `json:",optional"`
		KeyFile  string `json:",optional"`
		Verbose  bool   `json:",optional"`
		MaxConns int    `json:",default=10000"`
		MaxBytes int64  `json:",default=1048576"`
		// milliseconds
		Timeout      int64         `json:",default=3000"`
		CpuThreshold int64         `json:",default=900,range=[0:1000]"`
		Signature    SignatureConf `json:",optional"`
	}
```

