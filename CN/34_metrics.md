# Metrics in Ethereum and related Blockchain systems

在本文中，我们主要总结一下Ethereum中常用的Metrics。


## Metrics in Ethereum

描述以太坊中性能的两个重要的metrics是TPS和transaction delay.


## Compare with current relational database

Blockchain system 并不是传统意义上的Database management system.

### Metrics in Blockchain

- TPS
- Delay time
- Network I/O
- 支持的用户数量：无限的

### Metrics in relational database

- TPS 每秒钟处理的transaction的数量
- QPS 每秒钟查询的transaction的数量
- IOPS 每秒钟IO的读取的的速率
- 支持的用户数量：有限的


## How to configure influxDB 2.0 and Grafana

Geth 使用 influxDB 来实时监控 Metrics 数据，同时使用 Grafana 接入 influxDB 来实现时序数据的可视化。

值得注意的是，在 Geth 的官方文档中，默认接入的是 influxDB 1.0，而 influxDB 在几年前已经全面升级为 2.0 版本。两个版本之间存在一定的兼容性问题，包括数据库结构，查询语句的不同。因此，如果你使用的是 influxDB 2.0 版本，请参考这个 [PR](https://github.com/ethereum/go-ethereum/pull/23194) 来配置 Metrics 相关的参数。

```
 --metrics.influxdb.bucket "Your bucket name" --metrics.influxdb.token "Your Token" --metrics.influxdb.organization "Your Org name" 
```

此外，官方文档中使用的 [Grafana Dashboard](https://grafana.com/grafana/dashboards/13877-single-geth-dashboard/) 模版因为 influxDB Query 语句兼容性的问题，无法从 influxDB 2.0 中拿到数据。但是我们仍然可以使用 Grafana Dashboard 提供的样式，通过手动更换每个组件对应的 Query 语句的方式来使用该 Dashboard。
