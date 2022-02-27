# Jaeger
Jaeger 是 Uber 开发并开源的一款分布式追踪系统，兼容 OpenTracing API，适用于以下场景：
- 分布式跟踪信息传递
- 分布式事务监控
- 问题分析
- 服务依赖性分析
- 性能优化

Jaeger 的全链路追踪功能主要由三个角色完成:
- client：负责全链路上各个调用点的计时、采样，并将 tracing 数据发往本地 agent。
- agent：负责收集 client 发来的 tracing 数据，并以 thrift 协议转发给 collector。
- collector：负责搜集所有 agent 上报的 tracing 数据，统一存储。

# MYSQL
在docker的mysql容器中，获取mysql ip，使用该ip进入管理页面 \
- `cat /etc/hosts`
![](img/mysql_01.png)
