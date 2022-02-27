#  `Prometheus`
Prometheus 是一款基于时序数据库的开源监控告警系统，基本原理是通过 HTTP 协议周期性抓取被监控服务的状态，任意服务只要提供对应的 HTTP 接口就可以接入监控。不需要任何 SDK 或者其他的集成过程，输出被监控服务信息的 HTTP 接口被叫做 exporter 。
- 支持多维数据模型(由度量名和键值对组成的时间序列数据)
- 支持 PromQL 查询语言，可以完成非常复杂的查询和分析，对图表展示和告警非常有意义
- 不依赖分布式存储，单点服务器也可以使用
- 支持 HTTP 协议主动拉取方式采集时间序列数据
- 支持 PushGateway 推送时间序列数据
- 支持服务发现和静态配置两种方式获取监控目标
- 支持接入 Grafana

文件映射（win10文件目录）
```yaml
    volumes:
      - /e/code/golang/go_zero_project/gonivinck/prometheus/prometheus.yml:/opt/bitnami/prometheus/conf/prometheus.yml  # 将 prometheus 配置文件挂载到容器里
```

重置`grafana`密码：`grafana-cli admin reset-admin-password admin`
