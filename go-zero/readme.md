# About
本目录主要是对源码的阅读及记录

目前主要是根据`go-zero`生成项目后，从读取配置文件到启动相应服务，从`api`到`rpc`的顺序进行源代码的阅读。

项目微服务实现参考：[带你十天轻松搞定 Go 微服务系列](https://juejin.cn/user/2348212566892574/posts)

文件目录说明
- `core`
  - `conf`             读取并写入配置
  - `mapping`       将文件中的配置项转换为`struct`结构

## 目录
- [user/api](readme/user_api_user.go.md)：从 `user.go` 的 `main` 方法入手
- [user/rpc](readme/user_rpc_user.go.md)：从 `user.go` 的 `main` 方法入手

## go 语言知识

### interface类型转换
判断某个值是否实现某个接口
```go
type Stringer interface {
    String() string
}
// 同时 sv 转换为 Stringer 接口
if sv, ok := v.(Stringer); ok {
    fmt.Printf("v implements String(): %s\n", sv.String()) // note: sv, not v
}
```
用下划线判断接口是否被实现 \
我们定义了一个接口(interface)、定义了一个结构体(struct)
```go
type Foo interface {
     Say()
}
type Dog struct {
}
// 判断Dog这个struct是否实现了Foo这个interface
var _ Foo = Dog{}
```
用来判断Dog是否实现了Foo, 用作类型断言，如果Dog没有实现Foo，则会报编译错误
### 下划线妙用
仅对包进行初始化，而不使用包中的其他功能：
```go
import (
	"fmt" // 导入并使用
	// 仅初始化，而不使用
	// 执行本段代码之前会先调用test/foo中的初始化函数(init)
	_ "test/foo" 
)
```
