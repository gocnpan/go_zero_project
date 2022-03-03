# About
本目录主要是对源码的阅读及记录

目前主要是根据`go-zero`生成项目后，从读取配置文件到启动相应服务，从`api`到`rpc`的顺序进行源代码的阅读。

项目微服务实现参考：[带你十天轻松搞定 Go 微服务系列](https://juejin.cn/user/2348212566892574/posts)

文件目录说明
- `core`
  - `conf`             读取并写入配置
  - `mapping`       将文件中的配置项转换为`struct`结构

## 目录
- [`user/api`入口](readme/user_api_user.go.md) 从 `user.go` 的 `main` 方法入手
- [`user/rpc`入口](readme/user_rpc_user.go.md) 从 `user.go` 的 `main` 方法入手