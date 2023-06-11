# 交易和区块的同步

在本章中，我们会像哲学家一样思考：数据结构/实例/对象/变量是从哪里来，又要到哪里去呢？

## 概述

 在前面的章节中，我们已经讨论了在以太坊中 Transactions 是从 Transaction Pool 中，被 Validator/Miner 们验证打包，最终被保存在区块链中。那么，接下来的问题是，Transaction 是怎么被进入到 Transaction Pool 中的呢？基于同样的思考方式，那么一个刚刚在某个节点被打包好的 Block，它又将怎么传输到区块链网络中的其他节点那里，并最终实现 Blockchain 长度是一致的呢？在本章中，我们就来探索一下，节点是如何发送和接收 Transaction 和 Block 的。

## How Geth syncs Transactions：同步交易状态

在前面的章节中，我们曾经提到，Geth 节点中最顶级的对象是 Node 类型，负责节点最高级别生命周期相关的操作，例如节点的启动以及关闭，节点数据库的打开和关闭，启动RPC监听。而更具体的管理业务生命周期(lifecycle)的函数，都是由后端 Service 实例 `Ethereum` 和 `LesEthereum` 来实现的。

定义在`eth/backend.go` 中的 `Ethereum` 提供了一个全节点的所有的服务包括：TxPool 交易池， Miner 模块，共识模块，API 服务，以及解析从 P2P 网络中获取的数据。`LesEthereum` 提供了轻节点对应的服务。由于轻节点所支持的功能相对较少，在这里我们不过多描述。`Ethereum` 结构体的定义如下所示。

```go
type Ethereum struct {
	config *ethconfig.Config
	txPool     *txpool.TxPool
	blockchain *core.BlockChain
	handler            *handler // 我们关注的核心对象
	ethDialCandidates  enode.Iterator
	snapDialCandidates enode.Iterator
	merger             *consensus.Merger
	chainDb ethdb.Database // Block chain database
	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager
	bloomRequests     chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer      *core.ChainIndexer             // Bloom indexer operating during block imports
	closeBloomHandler chan struct{}
	APIBackend *EthAPIBackend
	miner     *miner.Miner
	gasPrice  *big.Int
	etherbase common.Address
	networkID     uint64
	netRPCService *ethapi.NetAPI
	p2pServer *p2p.Server
	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and etherbase)
	shutdownTracker *shutdowncheck.ShutdownTracker // Tracks if and when the node has shutdown ungracefully
}
```

这里值得提醒一下，在 Geth 代码中，不少地方都使用 `backend` 这个变量名，来指代 `Ethereum`。但是，也存在一些代码中使用 `backend` 来指代 `ethapi.Backend` 接口。在这里，我们可以做一下区分，`Ethereum` 负责维护节点后端的生命周期的函数，例如 Miner 的开启与关闭。而`ethapi.Backend` 接口主要是提供对外的业务接口，例如查询区块和交易的状态。读者可以根据上下文来判断 `backend` 具体指代的对象。我们在 geth 是如何启动的一章中提到，`Ethereum` 是在 Geth 启动的实例化的。在实例化 `Ethereum` 的过程中，就会创建一个 `APIBackend *EthAPIBackend` 的成员变量，它就是`ethapi.Backend` 接口类型的。



## How Geth syncs Blocks：同步区块状态
