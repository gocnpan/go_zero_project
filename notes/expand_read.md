# 拓展
来源：https://go-zero.dev/cn/extended-reading.html

## 日志
日志打印的级别与常规不一样，logx 支持的打印日志级别有：
- `alert`
- `info`
- `error`
- `severe`
- `fatal`
- `slow`
- `stat`

可配置参数
```go
const (
    // 打印所有级别的日志
    InfoLevel = iota
    // 打印 errors, slows, stacks 日志
    ErrorLevel
    // 仅打印 severe 级别日志
    SevereLevel
)
```  

模式：
- 一种文件输出：适合直接部署
- 一种控制台输出：k8s、docker等部署（由日志收集器导入至 es 进行分析）

禁用日志输出，将无法在次打开

## 布隆过滤器bloom
布隆过滤器，可以判断某元素在不在集合里面,因为存在一定的误判和删除复杂问题,一般的使用场景是:防止缓存击穿(防止恶意攻击)、 垃圾邮箱过滤、cache digests 、模型检测器等、判断是否存在某行数据,用以减少对磁盘访问，提高服务的访问性能。

## executors
executors 充当任务池，做多任务缓冲，适用于做批量处理的任务。如：clickhouse 大批量 insert，sql batch insert。


