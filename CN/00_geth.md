# [Chapter_00] 万物的起点: Geth Start

本章概要:

 1. `go-ethereum` 代码库的主要目录结构。
 2. `geth` 客户端/节点是如何启动的。
 3. 如何修改/添加 `geth` 对外的APIs。

## 什么是 geth？

`geth` 是以太坊基金会基于 Go 语言开发以太坊的官方客户端，它实现了 Ethereum 协议(黄皮书)中所有需要的实现的功能模块。我们可以通过启动 `geth` 来运行一个 Ethereum 的节点。在以太坊 Merge 之后，`geth` 作为节点的执行层继续在以太坊生态中发挥重要的作用。 `go-ethereum`是包含了 `geth` 客户端代码和以及编译 `geth` 所需要的其他代码在内的一个完整的代码库。在本系列中我们会通过深入 go-ethereum 代码库，从High-level 的 API 接口出发，沿着 Ethereum 主 Workflow，逐一的理解 Ethereum 具体实现的细节。

为了方便区分，在接下来的文章中，我们用 `geth` 来表示 Geth 客户端程序，用 go-ethereum (`Geth`)来表示 go-ethereum 的代码库。

总结的来说:

1. 基于 `go-ethereum` 代码库中的代码，我们可以编译出 `geth` 客户端程序。
2. 通过运行 `geth` 客户端程序我们可以启动一个 Ethereum 的节点。

### go-ethereum Codebase 结构

为了更好的从整体工作流的角度来理解 Ethereum，根据主要的业务功能，我们可以把 `go-ethereum` 划分成如下几个模块。

- Geth Client 模块
- Core 数据结构模块
- State Management 模块
  - StateDB 模块
  - Trie 数据结构模块
  - State Optimization (Pruning)
- Mining 模块
- EVM 模块
- P2P 网络模块
  - 节点数据同步
    - 交易数据
    - 区块数据
    - 区块链数据
- Storage 模块
  - 抽象数据库层
  - LevelDB 调用
- ...

目前，go-ethereum 代码库中的主要目录结构如下所示:

```
cmd/ 以太坊基金会官方开发的一些 Command-line 程序。该目录下的每个子目录都是一个单独运行的 CLI 程序。
   |── clef/ 以太坊官方推出的账户管理程序.
   |── geth/ 以太坊官方的节点客户端。
core/   以太坊核心模块，包括核心数据结构，statedb，EVM 等核心数据结构以及算法实现
   |── rawdb/ db 相关函数的高层封装(在 ethdb 和更底层的 leveldb 之上的封装)
      ├──accessors_state.go 从 Disk Level 读取/写入与 State 相关的数据结构。
   |── state/
      ├── statedb.go  StateDB 是管理以太坊 World State 最核心的代码，用于管理链上所有的 State 相关操作。
      ├── state_object.go state_object 是以太坊账户(包括 EOA & Contract)在 StateDB 具体的实现。
   |── txpool        Transaction Pool 相关的代码。
      |── txpool.go  Transaction Pool 的具体实现。
   |── types/  以太坊中最核心的数据结构
      |── block.go   以太坊 Block 的的数据结构定义与相关函数实现
      |── bloom9.go  以太坊使用的一个 Bloom Filter 的实现
      |── transaction.go 以太坊 Transaction 的数据结构定义与相关函数实现。
      |── transaction_signing.go 用于对 Transaction 进行签名的函数的实现。
      |── receipt.go  以太坊交易收据的实现，用于记录以太坊 Transaction 执行的结果
   |── vm/            以太坊的核心中核心 EVM 相关的一些的数据结构的定义。
      |── evm.go            EVM 数据结构和方法的定义
      |── instructions.go   EVM 指令的具体的定义，核心中的核心中的核心文件。
      |── logger.go   用于追踪 EVM 执行交易过程的日志接口的定义。具体的实现在eth/tracers/logger/logger.go 文件中。
      |── opcode.go   EVM 指令和数值的对应关系。
   |── genesis.go     创世区块相关的函数。每个 geth 客户端/以太坊节点初始化的都需要调用这个模块。
   |── state_processor.go EVM 执行交易的核心代码模块。 
console/
   |── bridge.go
   |── console.go  Geth Web3 控制台的入口
eth/      Ethereum 节点/后端/客户端具体功能定义和实现。例如节点的启动关闭，P2P 网络中交易和区块的同步。
ethdb/    Ethereum 本地存储的相关实现, 包括 leveldb 的调用
   |── leveldb/   Go-Ethereum使用的与 Bitcoin Core version一样的Leveldb作为本机存储用的数据库
internal/ 一些内部使用的工具库的集合，比如在测试用例中模拟 cmd 的工具。在构建 Ethereum 生态相关的工具时值得注意这个文件夹。
miner/
   |── miner.go   矿工模块的实现。
   |── worker.go  Block generation 的实现，包括打包 transaction，计算合法的 Block
p2p/     Ethereum 的P2P模块
   |── params    Ethereum 的一些参数的配置，例如: bootnode 的 enode 地址
   |── bootnodes.go  bootnode 的 enode 地址 like: aws 的一些节点，azure 的一些节点，Ethereum Foundation 的节点和 Rinkeby 测试网的节点
rlp/     RLP的 Encode与 Decode的相关
rpc/     Ethereum RPC客户端的实现
les/     Ethereum light client 轻节点的实现
trie/    Ethereum 中至关重要的数据结构 Merkle Patrica Trie(MPT) 的实现
   |── committer.go    Trie 向 Memory Database 提交数据的工具函数。
   |── database.go     Memory Database，是 Trie 数据和 Disk Database 提交的中间层。同时还实现了 Trie 剪枝的功能。**非常重要**
   |── node.go         MPT中的节点的定义以及相关的函数。
   |── secure_trie.go  基于 Trie 的封装的结构。与 trie 中的函数功能相同，不过secure_trie中的 key 是经过hashKey()函数hash过的，无法通过路径获得原始的 key值 
   |── stack_trie.go   Block 中使用的 Transaction/Receipt Trie 的实现
   |── trie.go         MPT 具体功能的函数实现。
 ```

## Geth Start

### 前奏: Geth Console

当我们想要部署一个 Ethereum 节点的时候，最直接的方式就是下载官方提供的发行版的 geth 客户端程序。`geth`是一个基于 CLI 的应用，启动`geth`和调用 `geth` 的功能性 API 需要使用对应的指令来操作。`geth` 提供了一个相对友好的 console 来方便用户调用各种指令。当我第一次阅读 Ethereum 的文档的时候，我曾经有过这样的疑问，为什么`geth`是由 Go 语言编写的，但是在官方文档中的 Web3 的API却是基于 Javascript 的调用？

这是因为 `geth` 内置了一个 Javascript 的解释器: *Goja* (interpreter)，来作为用户与 `geth` 交互的 CLI Console。我们可以在`console/console.go` 中找到它的定义。

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

### geth 节点是如何启动的

了解 Ethereum，我们首先要了解 Ethereum 客户端 Geth 是怎么运行的。 geth 程序的启动点位于 `cmd/geth/main.go/main()` 函数处，如下所示。

```go
func main() {
 if err := app.Run(os.Args); err != nil {
  fmt.Fprintln(os.Stderr, err)
  os.Exit(1)
 }
}
```

我们可以看到 `main()` 函数非常的简短，其主要功能就是启动一个解析 command line命令的工具: `gopkg.in/urfave/cli.v1`。继续深入，我们会发现在 cli app 初始化的时候会调用 `app.Action = geth` ，来调用 `geth()` 函数。而 `geth()` 函数就是用于启动 Ethereum 节点的顶层函数，其代码如下所示。

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

在 `geth()` 函数中，有三个比较重要的函数调用，分别是：`prepare()`，`makeFullNode()`，以及 `startNode()`。

`prepare()` 函数的实现就在当前的 `main.go` 文件中。它主要用于设置一些节点初始化需要的配置。比如，我们在节点启动时看到的这句话: *Starting Geth on Ethereum mainnet...* 就是在 `prepare()` 函数中被打印出来的。

`makeFullNode()` 函数的实现位于 `cmd/geth/config.go` 文件中。它会将 Geth 启动时的命令的上下文加载到配置中，并生成 `stack` 和`backend` 这两个实例。其中 `stack` 是一个 Node 类型的实例，它是通过 `makeFullNode()` 函数调用 `makeConfigNode()` 函数来初始化的。`Node` 是 geth 生命周期中最顶级的实例，它负责管理节点中的 P2P Server, Http Server, Database 等业务非直接相关的高级抽象。关于 Node 类型的定义位于`node/node.go`文件中。

这里的 `backend` 是一个 `ethapi.Backend` 类型的接口，提供了获取以太坊执行层运行时，所需要的基本函数功能。它的定义位于 `internal/ethapi/backend.go` 中。 由于这个接口中函数较多，我们选取了其中的部分关键函数方便大家理解这个接口所提供的基本功能，如下所示。

```go
type Backend interface {
	SyncProgress() ethereum.SyncProgress
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	ChainDb() ethdb.Database
	AccountManager() *accounts.Manager
	ExtRPCEnabled() bool
	RPCGasCap() uint64            // global gas cap for eth_call over rpc: DoS protection
	RPCEVMTimeout() time.Duration // global timeout for eth_call over rpc: DoS protection
	RPCTxFeeCap() float64         // global tx fee cap for all transaction related APIs
	UnprotectedAllowed() bool     // allows only for EIP155 transactions.
	SetHead(number uint64)
	HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, error)
	HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
	HeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Header, error)
	CurrentHeader() *types.Header
	CurrentBlock() *types.Header
	BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Block, error)
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
	BlockByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Block, error)
	StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *types.Header, error)
	StateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.StateDB, *types.Header, error)
	PendingBlockAndReceipts() (*types.Block, types.Receipts)
	GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error)
	GetTd(ctx context.Context, hash common.Hash) *big.Int
	GetEVM(ctx context.Context, msg *core.Message, state *state.StateDB, header *types.Header, vmConfig *vm.Config) (*vm.EVM, func() error, error)
	SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription
	SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription
	SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription
	SendTx(ctx context.Context, signedTx *types.Transaction) error
	GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, common.Hash, uint64, uint64, error)
	GetPoolTransactions() (types.Transactions, error)
	GetPoolTransaction(txHash common.Hash) *types.Transaction
	GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error)
	Stats() (pending int, queued int)
	TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions)
	TxPoolContentFrom(addr common.Address) (types.Transactions, types.Transactions)
	SubscribeNewTxsEvent(chan<- core.NewTxsEvent) event.Subscription
	ChainConfig() *params.ChainConfig
	Engine() consensus.Engine
	GetBody(ctx context.Context, hash common.Hash, number rpc.BlockNumber) (*types.Body, error)
	GetLogs(ctx context.Context, blockHash common.Hash, number uint64) ([][]*types.Log, error)
	SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription
	SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription
	SubscribePendingLogsEvent(ch chan<- []*types.Log) event.Subscription
	BloomStatus() (uint64, uint64)
	ServiceFilter(ctx context.Context, session *bloombits.MatcherSession)
}
```

我们可以发现 `ethapi.Backend` 接口主要对外提供了: 1. General Ethereum APIs, 这些 General APIs 对外提供了查询区块链节点管理对象的接口，例如 `ChainDb()` 返回当前节点的 DB 实例, `AccountManager()`; 2.  Blockchain 相关的 APIs, 例如链上数据的查询(Block & Transaction), `CurrentHeader(), BlockByNumber(), GetTransaction()`; 3. Transaction Pool 相关的APIs, 例如发送交易到本节点的 Transaction Pool, 以及查询交易池中的 Transactions, `GetPoolTransaction`。

目前 Geth 代码库中，有两个 `ethapi.Backend` 接口的实现，分别是: 1. 位于 `eth\api_backend` 中的 `EthAPIBackend`; 2. 位于 `les\api_backend` 的 `LesApiBackend`; 顾名思义，`EthAPIBackend` 提供了针对全节点的 Backend API 服务, 而 `LesApiBackend` 提供了轻节点的 Backend API 服务。总结的来说，如果读者想定制一些新的 RPC API，可以在 `ethapi.Backend` 接口中定义函数，并给 `EthAPIBackend` 添加具体的实现。

读者可能会发现，`ethapi.Backend` 接口所提供的函数功能，主要读写本地的维护的数据结构(i.e. Transaction Pool, Blockchain)的为主。那么作为一个有网络连接的 Backend, 以太坊的 Backend 或者说 Node 是怎么管理以太坊执行层节点的网络连接，共识等功能模块的呢？

我们深入 `makeFullNode()` 函数可以发现，生成`ethapi.Backend` 接口的语句 `backend, eth := utils.RegisterEthService(stack, &cfg.Eth)`, 还返回了另一个 `Ethereum` 类型的实例 `eth`。 这个 `Ethereum` 类型才是以太坊节点数结构中核心中的核心，它实现了以太坊全节点所需要的所有的 Service。它负责提供更为具体的以太坊的功能性 Service, 负责与以太坊业务直接相关的抽象，比如维护 Blockchain 的更新，共识算法，从 P2P 网络中同步区块，同步P2P节点远端的交易并放到交易池中，等业务功能。我们会在后续详细讲解 `Ethereum` 类型具体提供的服务。

`Ethereum` 实例根据上下文的配置信息在调用 `utils.RegisterEthService()` 函数生成。在`utils.RegisterEthService()`函数中，首先会根据当前的config来判断需要生成的Ethereum backend 的类型，是 light node backend 还是 full node backend。我们可以在 `eth/backend/new()` 函数和 `les/client.go/new()` 中找到这两种 Ethereum backend 的实例是如何初始化的。Ethereum backend 的实例定义了一些更底层的配置，比如chainid，链使用的共识算法的类型等。这两种后端服务的一个典型的区别是 light node backend 不能启动 Mining 服务。在 `utils.RegisterEthService()` 函数的最后，调用了 `Nodes.RegisterAPIs()` 函数，将刚刚生成的 backend 实例注册到 `stack` 实例中。

总结的说，`api_backend` 主要是用于对外提供查询，或者与后端功能性生命周期无关的函数，`Ethereum` 这类的节点层的后端，主要用于管理/控制节点后端的生命周期。

最后一个关键函数，`startNode()` 的作用是正式的启动一个以太坊执行层的节点。它通过调用 `utils.StartNode()` 函数来触发 `Node.Start()` 函数来启动`Stack`实例(Node)。在 `Node.Start()` 函数中，会遍历 `Node.lifecycles` 中注册的后端实例，并启动它们。此外，在 `startNode()` 函数中，还是调用了`unlockAccounts()` 函数，并将解锁的钱包注册到 `stack` 中，以及通过 `stack.Attach()` 函数创建了与 local Geth 交互的 RPClient 模块。

在 `geth()` 函数的最后，函数通过执行 `stack.Wait()`，使得主线程进入了阻塞状态，其他的功能模块的服务被分散到其他的子协程中进行维护。

### Node

正如我们前面提到的，Node 类型在 geth 的生命周期性中属于顶级实例，它负责作为与外部通信的高级抽象模块的管理员，比如管理 rpc server，http server，Web Socket，以及 P2P Server外 部接口。同时，Node 中维护了节点运行所需要的后端的实例和服务 (`lifecycles  []Lifecycle`)，例如我们上面提到的负责具体 Service 的`Ethereum` 类型。

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

在前面我们提到，整个程序的主线程因为调用了 `stack.Wait()` 而进入了阻塞状态。我们可以看到 Node 结构中声明了一个叫做 `stop` 的 channel。由于这个 Channel 一直没有被赋值，所以整个 geth 的主进程才进入了阻塞状态，持续并发的执行其他的业务协程。

```go
// Wait blocks until the node is closed.
func (n *Node) Wait() {
 <-n.stop
}
```

当 `n.stop` 这个 Channel 被赋予值的时候，`geth` 主函数就会停止当前的阻塞状态，并开始执行相应的一系列的资源释放的操作。这个地方的写法还是非常有意思的，值得我们参考。我们为读者编写了一个简单的示例:如何使用 Channel 来管理 Go 程序的生命周期。

值得注意的是，在目前的 go-ethereum 的 codebase 中，并没有直接通过给 `stop` 这个 channel 赋值方式来结束主进程的阻塞状态，而是使用一种更简洁粗暴的方式: 调用 `close()` 函数直接关闭 Channel。我们可以在 `node.doClose()` 找到相关的实现。`close()` 是 go 语言的原生函数，用于关闭 Channel 时使用。

```go
// doClose releases resources acquired by New(), collecting errors.
func (n *Node) doClose(errs []error) error {
 // Close databases. This needs the lock because it needs to
 // synchronize with OpenDatabase*.
 n.lock.Lock()
 n.state = closedState
 errs = append(errs, n.closeDatabases()...)
 n.lock.Unlock()

 if err := n.accman.Close(); err != nil {
  errs = append(errs, err)
 }
 if n.keyDirTemp {
  if err := os.RemoveAll(n.keyDir); err != nil {
   errs = append(errs, err)
  }
 }

 // Release instance directory lock.
 n.closeDataDir()

 // Unblock n.Wait.
 close(n.stop)

 // Report any errors that might have occurred.
 switch len(errs) {
 case 0:
  return nil
 case 1:
  return errs[0]
 default:
  return fmt.Errorf("%v", errs)
 }
}
```

### Ethereum Backend

我们可以在 `eth/backend.go` 中找到 `Ethereum` 这个结构体的定义。这个结构体包含的成员变量以及接收的方法实现了一个 Ethereum full node 所需要的全部功能和数据结构。我们可以在下面的代码定义中看到，Ethereum结构体中包含 `TxPool`，`Blockchain`，`consensus.Engine`，`miner`等最核心的几个数据结构作为成员变量，我们会在后面的章节中详细的讲述这些核心数据结构的主要功能，以及它们的实现的方法。

```go
type Ethereum struct {
	config *ethconfig.Config
	txPool     *txpool.TxPool
	blockchain *core.BlockChain
   handler            *handler     // handler 是P2P 网络数据同步的核心实例，我们会在后续的网络同步模块仔细的讲解它的功能
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

节点启动和停止 Mining 的就是通过调用 `Ethereum.StartMining()` 和 `Ethereum.StopMining()` 实现的。设置 Mining 的收益账户是通过调用 `Ethereum.SetEtherbase()` 实现的。

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

这里我们额外关注一下 `handler` 这个成员变量。`handler` 的定义在 `eth/handler.go` 中。

我们从从宏观角度来看，一个节点的主工作流需要: 1.从网络中获取/同步 Transaction 和 Block 的数据 2. 将网络中获取到 Block 添加到 Blockchain 中。而 `handler` 就负责提供中同步区块和交易数据的功能，例如，`downloader.Downloader` 负责从网络中同步 Block ，`fetcher.TxFetcher` 负责从网络中同步交易。关于这些方法的具体实现，我们会在后续章节：数据同步中详细介绍。

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

到此，我们就介绍了 geth 及其所需要的基本模块是如何启动的和关闭的。我们在接下来将视角转入到各个模块中，从更细粒度的角度深入探索 Ethereum 的具体实现。

### Related Terms

- Geth
- go-ethereum (geth)

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
