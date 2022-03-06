文件：[rest/server.go](../rest/server.go)

## 简介
1. 生成 `http server` 结构体：包含服务引擎、路由字典
2. 添加路由、中间件、开启服务等功能

## 详细
### `Server` 结构体
```go
	Server struct {
		// 服务 核心？中间数据？
		// 记录内容：路由、handler、middleware等内容
		ngin   *engine
		// http 服务构建者
		router httpx.Router // router.NewRouter()，来自rest/router/patrouter.go
	}
```
通过`NewServer`方法生成

### 路由配置
`server.AddRoutes`方法，按配置方法的种类分别添加路由信息，举例
```go
	// 带 jwt 权限校验
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/api/user/userinfo",
				Handler: UserInfoHandler(serverCtx),
			},
		},
		// 配置方法
		// 为请求附加可用jwt & 密钥信息
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
	)
```
在 [server.go](../rest/server.go) 文件中，`AddRoutes`方法：
- 运行：配置 featuredRoutes 参数的方法，为当前路由设置相应的参数
  - 如`rest.WithJwt`方法：
    ```go
    func WithJwt(secret string) RouteOption {
        return func(r *featuredRoutes) {
            validateSecret(secret)
            r.jwt.enabled = true
            r.jwt.secret = secret
        }
    }
    ```
  
### `Start`
通过 `bindRoutes --> ... --> bindRoute` 方法，处理中间件、handler
- 会绑定预设的配置方法：如链路跟踪、降载、解压等中间件
- 同时绑定用户中间件
- handler方法以中间件形式最后插入

`bindRoute`调用`rest/router/patrouter.go`中的 `Handle`方法，将请求路径、处理方法填入字典
- 按请求方法分类填入
- 搜索树类的数据

没有`ssl`证书的情况下，开启`http`服务，否则开启`https`服务
- 开启服务的详细代码：[rest/internal/starter.go](../rest/internal/starter.go)

## 补充
### Prometheus 微服务监控
在请求服务端时，监控指标数据不断被Prometheus 中间件搜集。
1. HistogramVec 负责请求耗时搜集：
   - bucket 存放的就是 option 指定的耗时指标。某个请求耗时多少就会被聚集对应的桶，计数。
   - 最终展示的就是一个路由在不同耗时的分布，很直观提供给开发者可以优化的区域。


2. CounterVec 负责指定 labels 标签搜集：
   - Labels: []string{"path", "code"}
   - labels 相当一个 tuple。go-zero 是以(path, code)作为整体，记录不同路由不同状态码的返回次数。如果 4xx,5xx过多的时候，是不是应该看看你的服务健康程度？




