# 00_万物的起点: Geth Start \!

## 什么是Geth？

Geth是基于Go语言开发以太坊的客户端，它实现了Ethereum协议(黄皮书)中所有需要的实现的功能模块，包括状态管理，挖矿，P2P网络通信，密码学，数据库，EVM解释器等。我们可以通过启动Geth来运行一个Ethereum的节点。Go-ethereum是包含了Geth在内的一个代码库，它包含了Geth，以及编译Geth所需要的其他代码。在本系列中，我们会深入Go-ethereum代码库，从High-level的API接口出发，沿着Ethereum主Workflow，从而理解Ethereum具体实现的细节。

### Go-ethereum Codebase 结构

为了更好的从整体工作流的角度来理解Ethereum，根据主要的业务功能，我们将go-ethereum划分成如下几个模块来分析。

- Geth Client模块
- Core数据结构模块
- State Management模块
  - StateDB 模块
  - Trie 模块
  - State Optimization (Pruning)
- Mining模块
- EVM 模块
- P2P 网络模块
  - 节点数据同步
- ...

目前，go-ethereum项目的主要目录结构如下所示:

```
cmd/ ethereum相关的Command-line程序。该目录下的每个子目录都包含一个可运行的main.go。
   |── clef/ Ethereum官方推出的Account管理程序.
   |── geth/ Geth的本体。
core/   以太坊核心模块，包括核心数据结构，statedb，EVM等算法实现
   |── rawdb/ db相关函数的高层封装(在ethdb和更底层的leveldb之上的封装)
   |── state/
       ├──statedb.go  StateDB结构用于存储所有的与Merkle trie相关的存储, 包括一些循环state结构  
   |── types/  包括Block在内的以太坊核心数据结构
      |── block.go  以太坊block
      |── bloom9.go  一个Bloom Filter的实现
      |── transaction.go 以太坊transaction的数据结构与实现
      |── transaction_signing.go 用于对transaction进行签名的函数的实现
      |── receipt.go  以太坊收据的实现，用于说明以太坊交易的结果
   |── vm/
   |── genesis.go     创世区块相关的函数，在每个geth初始化的都需要调用这个模块
   |── tx_pool.go     Ethereum Transaction Pool的实现
consensus/
   |── consensus.go   共识相关的参数设定，包括Block Reward的数量
console/
   |── bridge.go
   |── console.go  Geth Web3 控制台的入口
ethdb/    Ethereum 本地存储的相关实现, 包括leveldb的调用
   |── leveldb/   Go-Ethereum使用的与Bitcoin Core version一样的Leveldb作为本机存储用的数据库
miner/
   |── miner.go   矿工模块的实现。
   |── worker.go  真正的block generation的实现实现，包括打包transaction，计算合法的Block
p2p/     Ethereum 的P2P模块
   |── params    Ethereum 的一些参数的配置，例如: bootnode的enode地址
   |── bootnodes.go  bootnode的enode地址 like: aws的一些节点，azure的一些节点，Ethereum Foundation的节点和 Rinkeby测试网的节点
rlp/     RLP的Encode与Decode的相关
rpc/     Ethereum RPC客户端的实现
les/     Ethereum light client的实现
trie/    Ethereum 中至关重要的数据结构 Merkle Patrica Trie(MPT)的实现
   |── committer.go    Trie向Memory Database提交数据的工具函数。
   |── database.go     Memory Database，是Trie数据和Disk Database提交的中间层。同时还实现了Trie剪枝的功能。**非常重要**
   |── node.go         MPT中的节点的定义以及相关的函数。
   |── secure_trie.go  基于Trie的封装的Trie结构。与trie中的函数功能相同，不过secure_trie中的key是经过hashKey()函数hash过的，无法通过路径获得原始的key值
   |── stack_trie.go   Block中使用的Transaction/Receipt Trie的实现
   |── trie.go         MPT具体功能的函数实现
 ```

## Geth Start

### 前奏: Geth Console

当我们想要部署一个Ethereum节点的时候，最直接的方式就是下载官方提供的发行版的geth程序。Geth是一个基于CLI的应用，启动Geth和调用Geth的功能性API需要使用对应的指令来操作。Geth提供了一个相对友好的console来方便用户调用各种指令。当我第一次阅读Ethereum的文档的时候，我曾经有过这样的疑问，为什么Geth是由Go语言编写的，但是在官方文档中的Web3的API却是基于Javascript的调用？

这是因为Geth内置了一个Javascript的解释器:å*Goja* (interpreter)，来作为用户与Geth交互的CLI Console。我们可以在`console/console.go`中找到它的定义。

<!-- /*Goja is an implementation of ECMAScript 5.1 in Pure GO*/ -->

```go
// Console is a JavaScript interpreted runtime environment. It is a fully fledged
// JavaScript console attached to a running node via an external or in-process RPC
// client.
type Console struct {
 client   *rpc.Client         // RPC client to execute Ethereum requests through
 jsre     *jsre.JSRE          // JavaScript runtime environment running the interpreter
 prompt   string              // Input prompt prefix string
 prompter prompt.UserPrompter // Input prompter to allow interactive user feedback
 histPath string              // Absolute path to the console scrollback history
 history  []string            // Scroll history maintained by the console
 printer  io.Writer           // Output writer to serialize any display strings to
}
```

### 启动

了解Ethereum，我们首先要了解Ethereum客户端Geth是怎么运行的。

 <!-- `geth console 2` -->

Geth的启动点位于`cmd/geth/main.go/main()`函数处，如下所示。

```go
func main() {
 if err := app.Run(os.Args); err != nil {
  fmt.Fprintln(os.Stderr, err)
  os.Exit(1)
 }
}
```

我们可以看到`main()`函数非常的简短，其主要功能就是启动一个解析 command line命令的工具: `gopkg.in/urfave/cli.v1`。我们会发现在cli app初始化的时候会调用`app.Action = geth`，来调用`geth()`函数。`geth()`函数就是用于启动Ethereum节点的顶层函数，其代码如下所示。

```go
func geth(ctx *cli.Context) error {
 if args := ctx.Args(); len(args) > 0 {
  return fmt.Errorf("invalid command: %q", args[0])
 }

 prepare(ctx)
 stack, backend := makeFullNode(ctx)
 defer stack.Close()

 startNode(ctx, stack, backend, false)
 stack.Wait()
 return nil
}
```

在`geth()`函数，我们可以看到有三个比较重要的函数调用，分别是：`prepare()`，`makeFullNode()`，以及`startNode()`。

`prepare()` 函数的实现就在当前的`main.go`文件中。它主要用于设置一些节点初始化需要的配置。比如，我们在节点启动时看到的这句话: *Starting Geth on Ethereum mainnet...* 就是在`prepare()`函数中被打印出来的。

`makeFullNode()`函数的实现位于`cmd/geth/config.go`文件中。它会将Geth启动时的命令的上下文加载到配置中，并生成`stack`和`backend`这两个实例。其中`stack`是一个Node类型的实例，它是通过`makeFullNode()`函数调用`makeConfigNode()`函数来生成。Node是Geth生命周期中最顶级的实例，它的开启和关闭与Geth的启动和关闭直接对应。关于Node类型的定义位于`node/node.go`文件中。

`backend`实例是指的是具体Ethereum Client的功能性实例。它是一个Ethereum类型的实例，负责提供更为具体的以太坊的功能性Service，比如管理Blockchain，共识算法等具体模块。它根据上下文的配置信息在调用`utils.RegisterEthService()`函数生成。在`utils.RegisterEthService()`函数中，首先会根据当前的config来判断需要生成的Ethereum backend的类型，是light node backend还是full node backend。我们可以在`eth/backend/new()`函数和`les/client.go/new()`中找到这两种Ethereum backend的实例是如何初始化的。Ethereum backend的实例定义了一些更底层的配置，比如chainid，链使用的共识算法的类型等。这两种后端服务的一个典型的区别是light node backend不能启动Mining服务。在`utils.RegisterEthService()`函数的最后，调用了`Nodes.RegisterAPIs()`函数，将刚刚生成的backend实例注册到`stack`实例中。

```go
 eth := &Ethereum{
  config:            config,
  merger:            merger,
  chainDb:           chainDb,
  eventMux:          stack.EventMux(),
  accountManager:    stack.AccountManager(),
  engine:            ethconfig.CreateConsensusEngine(stack, chainConfig, &ethashConfig, config.Miner.Notify, config.Miner.Noverify, chainDb),
  closeBloomHandler: make(chan struct{}),
  networkID:         config.NetworkId,
  gasPrice:          config.Miner.GasPrice,
  etherbase:         config.Miner.Etherbase,
  bloomRequests:     make(chan chan *bloombits.Retrieval),
  bloomIndexer:      core.NewBloomIndexer(chainDb, params.BloomBitsBlocks, params.BloomConfirms),
  p2pServer:         stack.Server(),
  shutdownTracker:   shutdowncheck.NewShutdownTracker(chainDb),
 }
```

`startNode()`函数的作用是正式的启动一个Ethereum Node。它通过调用`utils.StartNode()`函数来触发`Node.Start()`函数来启动`Stack`实例（Node）。在`Node.Start()`函数中，会遍历`Node.lifecycles`中注册的后端实例，并在启动它们。此外，在`startNode()`函数中，还是调用了`unlockAccounts()`函数，并将解锁的钱包注册到`stack`中，以及通过`stack.Attach()`函数创建了与local Geth交互的RPClient模块。

在`geth()`函数的最后，函数通过执行`stack.Wait()`，使得主线程进入了阻塞状态，其他的功能模块的服务被分散到其他的子协程中进行维护。

### Node

正如我们前面提到的，Node类型在Geth的生命周期性中属于顶级实例，它负责作为与外部通信的外部接口，比如管理rpc server，http server，Web Socket，以及P2P Server外部接口。同时，Node中维护了节点运行所需要的后端的实例和服务(`lifecycles  []Lifecycle`)。

```go
// Node is a container on which services can be registered.
type Node struct {
 eventmux      *event.TypeMux
 config        *Config
 accman        *accounts.Manager
 log           log.Logger
 keyDir        string            // key store directory
 keyDirTemp    bool              // If true, key directory will be removed by Stop
 dirLock       fileutil.Releaser // prevents concurrent use of instance directory
 stop          chan struct{}     // Channel to wait for termination notifications
 server        *p2p.Server       // Currently running P2P networking layer
 startStopLock sync.Mutex        // Start/Stop are protected by an additional lock
 state         int               // Tracks state of node lifecycle

 lock          sync.Mutex
 lifecycles    []Lifecycle // All registered backends, services, and auxiliary services that have a lifecycle
 rpcAPIs       []rpc.API   // List of APIs currently provided by the node
 http          *httpServer //
 ws            *httpServer //
 httpAuth      *httpServer //
 wsAuth        *httpServer //
 ipc           *ipcServer  // Stores information about the ipc http server
 inprocHandler *rpc.Server // In-process RPC request handler to process the API requests

 databases map[*closeTrackingDB]struct{} // All open databases
}
```

#### Node的关闭

在前面我们提到，整个程序的主线程因为调用了`stack.Wait()`而进入了阻塞状态。我们可以看到Node结构中声明了一个叫做`stop`的channel。由于这个Channel一直没有被赋值，所以整个Geth的主进程才进入了阻塞状态，并不断循环的执行其他的业务协程。
```go
// Wait blocks until the node is closed.
func (n *Node) Wait() {
	<-n.stop
}
```


### Ethereum API Backend

我们可以在`eth/backend.go`中找到`Ethereum`这个结构体的定义。这个结构体包含的成员变量以及接收的方法实现了一个Ethereum full node所需要的全部功能和数据结构。我们可以在下面的代码定义中看到，Ethereum结构体中包含了`TxPool`，`Blockchain`，`consensus.Engine`，`miner`等最核心的几个数据结构作为成员变量，我们会在后面的章节中详细的讲述这些核心数据结构的主要功能，以及它们的实现的方法。

```go
// Ethereum implements the Ethereum full node service.
type Ethereum struct {
 config *ethconfig.Config

 // Handlers
 txPool             *core.TxPool
 blockchain         *core.BlockChain
 handler            *handler
 ethDialCandidates  enode.Iterator
 snapDialCandidates enode.Iterator
 merger             *consensus.Merger

 // DB interfaces
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
 netRPCService *ethapi.PublicNetAPI

 p2pServer *p2p.Server

 lock sync.RWMutex // Protects the variadic fields (e.g. gas price and etherbase)

 shutdownTracker *shutdowncheck.ShutdownTracker // Tracks if and when the node has shutdown ungracefully
}

```

节点启动和停止Mining的就是通过调用`Ethereum.StartMining()`和`Ethereum.StopMining()`实现的。设置Mining的收益账户是通过调用`Ethereum.SetEtherbase()`实现的。

```go
// StartMining starts the miner with the given number of CPU threads. If mining
// is already running, this method adjust the number of threads allowed to use
// and updates the minimum price required by the transaction pool.
func (s *Ethereum) StartMining(threads int) error {
   ...
 // If the miner was not running, initialize it
 if !s.IsMining() {
      ...
      // Start Mining
  go s.miner.Start(eb)
 }
 return nil
}
```

这里我们额外关注一下`handler`这个成员变量。`handler`的定义在`eth/handler.go`中。

我们从从宏观角度来看，一个节点的主工作流需要: 1.从网络中获取/同步Transaction和Block的数据 2. 将网络中获取到Block添加到Blockchain中。而`handler`就维护了backend中同步/请求数据的实例，比如`downloader.Downloader`，`fetcher.TxFetcher`。关于这些成员变量的具体实现，我们会在后续的文章中详细介绍。

```go
type handler struct {
 networkID  uint64
 forkFilter forkid.Filter // Fork ID filter, constant across the lifetime of the node

 snapSync  uint32 // Flag whether snap sync is enabled (gets disabled if we already have blocks)
 acceptTxs uint32 // Flag whether we're considered synchronised (enables transaction processing)

 checkpointNumber uint64      // Block number for the sync progress validator to cross reference
 checkpointHash   common.Hash // Block hash for the sync progress validator to cross reference

 database ethdb.Database
 txpool   txPool
 chain    *core.BlockChain
 maxPeers int

 downloader   *downloader.Downloader
 blockFetcher *fetcher.BlockFetcher
 txFetcher    *fetcher.TxFetcher
 peers        *peerSet
 merger       *consensus.Merger

 eventMux      *event.TypeMux
 txsCh         chan core.NewTxsEvent
 txsSub        event.Subscription
 minedBlockSub *event.TypeMuxSubscription

 peerRequiredBlocks map[uint64]common.Hash

 // channels for fetcher, syncer, txsyncLoop
 quitSync chan struct{}

 chainSync *chainSyncer
 wg        sync.WaitGroup
 peerWG    sync.WaitGroup
}
```

这样，我们就介绍了Geth及其所需要的基本模块是如何启动的。我们在接下来将视角转入到各个模块中，从更细粒度的角度深入Ethereum的实现。

### Appendix

这里补充一个Go语言的语法知识: **类型断言**。在`Ethereum.StartMining()`函数中，出现了`if c, ok := s.engine.(*clique.Clique); ok`的写法。这中写法是Golang中的语法糖，称为类型断言。具体的语法是`value, ok := element.(T)`，它的含义是如果`element`是`T`类型的话，那么ok等于`True`, `value`等于`element`的值。在`if c, ok := s.engine.(*clique.Clique); ok`语句中，就是在判断`s.engine`的是否为`*clique.Clique`类型。

```go
  var cli *clique.Clique
  if c, ok := s.engine.(*clique.Clique); ok {
   cli = c
  } else if cl, ok := s.engine.(*beacon.Beacon); ok {
   if c, ok := cl.InnerEngine().(*clique.Clique); ok {
    cli = c
   }
  }
```