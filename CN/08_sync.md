# [WIP] Get transactions and blocks from peers

## General

就像哲学中经常思考的问题：从哪里来，到哪里去。在以太坊中，我们同样也好奇：对于一个节点来说， Transaction, Block 是从哪里来的（怎样来的），又怎样被融合进 Blockchain 中的。在前面的章节中，我们分析了在 Geth 中，Transaction 和 Block 是如何构建的。在本章中，我们来探索一下，节点是如何发送和接收 Transaction 以及 Block 的。

## How Geth syncing Transactions

在前面的章节中，我们曾经提到，Geth 客户端或者说 Geth 节点中最顶级的对象是 Node 类型，负责节点生命周期相关的操作，例如节点的启动以及关闭，节点数据库的打开和关闭，启动RPC监听，注册后端API等服务。而更具体的服务逻辑，都是由后端 Service 实例 `Ethereum` 和 `LesEthereum` 来实现的。定义在`eth/backend.go` 中的 `Ethereum` 提供了一个全节点 Client 所需要的所有的服务，其中就包括：维护 TxPool 交易池，维护 Miner 模块，共识模块，API 服务，以及解析处理从 P2P 网络中获取的数据。


## How Geth syncing Blocks
