# 00_万物的起点从geth出发



## 什么是Geth？

Geth是基于Go语言开发以太坊的客户端，也是go-ethereum代码库编译来的可执行程序。Geth实现了Ethereum黄皮书中所有需要的实现的功能，包括状态管理，挖矿，网络通信，密码学模块，数据库模块，EVM解释器等模块。

Geth对用户提供了高层的API方便调用，我们需要的就是深入这些高层的API内部，了解Ethereum具体实现的细节。

### Geth CLI

当我们想要部署一个Ethereum节点的时候，最直接的方式就是下载官方提供的发行版的geth程序。一般情况下，geth表现出来的是一个基于CLI的应用。当我们想要使用geth的功能的时候，要配合对应的指令来操作。

当我第一次阅读Ethereum的文档的时候，我曾经有过这样的疑问，为什么Go语言写成的Ethereum，但是在官方文档中却描述了的是Javascript的API使用？

这是因为Geth 内置了一个Javascript的解释器Goja (interpreter)，来构造CLI Console方便与用户交互。

在console/console.go的代码中我们可以看到，geth中与用户交互的console，其实依赖于Geth代码中内置的Javascript的解释器，通过RPC请求来获取当前链上的信息，以及与链进行数据交互。

整个geth程序中，数据对外交互的窗口只有RPC接口提供的服务。

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

<!-- /*Goja is an implementation of ECMAScript 5.1 in Pure GO*/ -->

### Go-ethereum 代码目录结构

为了更好的从整体工作流的角度来理解Ethereum，根据主要的业务功能，我们将go-ethereum划分成如下几个模块来分析。

- Geth Client模块
- State Management模块
- Mining模块
- P2P 网络模块
- ...

了解Ethereum，我们首先要了解Ethereum客户端是怎么运行的。

 <!-- `geth console 2` -->

目前，go-ethereum项目的主要目录结构如下所示。

```
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
   |── miner.go   矿工的基本的实现。
   |── worker.go  真正的block generation的实现实现，包括打包transaction，计算合法的Block
p2p/     Ethereum 的P2P模块
   |── params    Ethereum 的一些参数的配置，例如: bootnode的enode地址
   |── bootnodes.go  bootnode的enode地址 like: aws的一些节点，azure的一些节点，Ethereum Foundation的节点和 Rinkeby测试网的节点
rlp/     RLP的Encode与Decode的相关
rpc/     Ethereum RPC客户端的实现
les/     Ethereum light client的实现
trie/    Ethereum 中至关重要的数据结构 Merkle Patrica Trie(MPT)的实现
   |── node.go         MPT中的节点的定义以及相关的函数。
   |── secure_trie.go  与trie中的函数功能相同，不过secure_trie中的key是经过hashKey()函数hash过的，无法通过路径获得原始的key值
   |── stack_trie.go   Block中使用的Transaction/Receipt Trie的实现
   |── trie.go MPT具体功能的函数实现
 ```