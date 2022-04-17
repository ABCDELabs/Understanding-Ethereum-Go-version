# Understanding-Ethereum-Go-version

- Title: Understanding Ethereum: Starting with go-ethereum Source Code｜理解以太坊: Go-Ethereum 源码剖析
- Subject: Ethereum Source Code Analysis | 以太坊源码剖析
- Author: Siyuan Han
- Go-Ethereum Version: v1.10.15 (London Upgrade)
- Updated date: 2022-04

## Preface

### Background

距中本聪发表比特币白皮书至今已经过了十几年的时光。Blockchain这一理念，从最开始作为支持Bitcoin的分布式账本系统出发，在这十几年的演化中不断发展，从专注于加密货币到如今的逐渐通用化，逐渐成为了集成了包括*数据库*，*分布式系统*，*点对点网络*，*编译原理*，*静态软件分析*，*众包*，*密码学*，*经济学*，*货币金融学*，*社会学*在内的等多个学科知识的一个全新技术领域。至今仍然是时下**最热度最高**，**最能引起社会讨论**的技术话题之一，逐渐从小众的去中心化社区逐渐走向了主流社会的舞台。

目前，市面上绝大多数的热门Blockchain系统都已经开源，并继续以开源的形式在社区的维护下持续开发中。这就为我们提供了一种的极好的学习Blockchain技术的方式: 通过结合文档，结合源代码的方式基于State-of-the-arts的Blockchain Systems出发开始学习。

一个热门的技术总是会有大量的文档，资料，经过不同视角的作者记录下来。不管是探究以加密货币导向（Crypto-based）的Bitcoin, 还是了解致力于实现通用Web3.0框架（General-Purpose）的Ethereum，有丰厚的文档从high-level的角度来讲述它们的基础概念，以及设计的思想。比如，技术社区有非常多的资料来讲述什么是梅克尔树 (Merkle Hash Tree)，什么是梅克尔帕特里夏树 (Merkle Patricia Tree)，什么是有向无环图 (Directed acyclic Graph); BFT (Byzantine Fault Tolerance)和 PoW (Proof-Of-Work) 共识算法算法的区别; 以及介绍Blockchain系统为什么可以抵抗双花攻击 (Double-Spending)，或者为什么Ethereum会遇到 DAO Attack (Decentralized autonomous organization) 等具体问题。

但是，只了解关键组件的实现细节，或者高度抽象的系统工作流，并不代表着可以让读者从真正搞清楚Blockchain的工作原理，反而很容易让读者在一些关键细节上一头雾水，似懂非懂。比如，当我们谈到Blockchain系统中Transaction的生命周期，我们在文档中经常会读到，Miner节点批量地从自己维护的Transaction pool中选择一些Transaction并打包成一个新的Block中。那么究竟miner是怎么从网络中获取到Transaction？又是继续什么样的选择策略从Transaction pool中选取Transaction？最终按照怎么的order把transaction打包进区块链中的呢？如何打包成功的Block是怎么交互给其他节点呢？我尝试去搜索了一下，发现鲜有文章从*整体*的系统工作流的角度出发，对以细粒度的视角对区块链系统中的具体的实现*细节*进行解析。与数据库系统(Database Management System)相似，Blockchain系统同样是一个包含网络层，业务逻辑层，任务解析层，存储层的复杂数据管理系统。对它研究同样需要从系统的实现细节出发，从宏观到微观的了解每个执行逻辑的工作流，才能彻底理解和掌握这门技术的秘密。

笔者坚信，随着网络基础架构的不断完善，将带来的显著的带宽上升和通信延迟下降。同时随着存储技术以及算法的不断发展，未来系统的运行效率将会不断逼近硬件极限。在未来的是五到十年内，云端服务/去中心化系统的效率以及覆盖场景一定还会有很大的提升。未来技术世界一定是两极分化的。一极是以大云计算公司（i.e, Google，MS，Oracle，Snowflake，and Alibaba）为代表的中心化服务商。另一极就是以Blockchain技术作为核心的去中心化的世界。在这个世界中，Ethereum及其生态系统是当之无愧的领头羊。Ethereum 不仅在作为Public Chain的自身取得了巨大的成功，构建了强大的生态系统，并且吸引一大批世界上最优秀的工程师，研究人员的持续的发力，不断的将新思想，新理念，新技术引入到Ethereum中，持续的引领Blockchain技术发展。而且Go-Ethereum作为其优秀的开源实现，已经被广泛的订制，来适应不同的私有/联盟场景(e.g., Quorum, Binance Smart Chain)。因此，要想真正掌握好区块链系统的原理，达到可以设计开发区块链系统的水平，研究好Ethereum的原理以及其设计思想是非常有必要。

本系列文章，作为我在博士期间学习/研究的记录，将会从Blockchain中具体业务的工作的视角出发，在源码的层面细粒度的解析以太坊系统中各个模块的实现的细节，以及背后的蕴含的技术和设计思想。同时，在阅读源代码中发现的问题也可以会提交Pr来贡献社区。本系列基于的代码库是Go-ethereum version 1.10.*。Go-ethereum是以太坊协议的Go语言实现版本，目前由以太坊基金会维护。目前除了Go-ethereum之外，Ethereum还有C++, Python，Java, Rust等基于其他语言实现的版本。相比于其他的社区版实现，go-ethereum的用户数量最多，开发人员最多，版本更新最频繁，issues的发现和处理都较快，运行也更更加的稳定。其他语言的Ethereum实现版本因为用户与开发人员的数量相对较少，更新频率相对较低，隐藏问题出现的可能性更高。因此我们选择从Go-ethereum代码库作为我们的主要学习资料，来理解Ethereum系统与网络的设计实现。

### 为什么要阅读区块链系统的源代码

1. 文档资料相对较少，且**内容浅尝辄止**，**相互独立**。比如，*很多的科普文章都提到，在打包新的Block的时候，miner负责把a batch of transactions从transaction pool中打包到新的block中*。那么我们希望读者思考如下的几个问题：
    - Miner是基于什么样策略从Transaction Pool中选择Transaction呢？
    - 被选择的Transactions又是以怎样的顺序(Order)被打包到区块中的呢？
    - 在执行Transaction的EVM是怎么计算gas used，从而限定Block中Transaction的数量的呢?
    - 剩余的gas又是怎么返还给Transaction Proposer的呢？
    - EVM是怎么解释Contract的Message Call并执行的呢？
    - 在执行Transaction中是哪个模块，又是怎样去修改Contract中的持久化变量呢？
    - Smart Contract中的持久化变量又是以什么样的形式存储的呢？
    - 当新的Block加入到Blockchain中时，World State又是何时怎样更新的呢？
    - 哪些数据常驻内存，哪些数据又需要保存在Disk中呢？

2. 目前的Blockchain系统并没有像数据库系统(DBMS)那样统一实现的方法论，每个不同的系统中都集成了大量的细节。从源码的角度出发可以了解到很多容易被忽视的细节。简单的说，一个完整的区块链系统至少包含以下的模块:
    - 密码学模块: 加解密，签名，安全hash，Mining
    - 网络模块: P2P节点通信
    - 分布式共识模块: PoW, BFT
    - 智能合约解释器模块: Solidity编译语言，EVM解释器
    - 数据存储模块: 数据库，数据存储，Index，LevelDB
    - Log日志模块
    - etc.

而在具体实现中，由于设计理念，以及go语言的特性(没有继承派生关系)，go-ethereum中的模块之间相互调用关系相对复杂。因此，只有通过阅读源码的方式才能更好理解不同模块之间的调用关系，以及业务的workflow的执行流程/细节。

### Blockchain System (BCS) VS Database Management System (DBMS)

Blockchain 系统在设计层面借鉴了很多数据库系统中的设计逻辑。

- Blockchain系统同样也从Transaction作为基本的操作载核，包含一个Parser模块，Transaction Executor模块，和一个Storage 管理模块。

## Contents(暂定)

### PART ONE - General Source Code Analysis: Basic Workflow and Data Components

- [00_万物的起点: Geth Start](CN/00_geth.md)
- [01_Account and State](CN/01_account_state.md)
- [02_State Management i: StateDB](CN/02_state_management_statedb.md)
- [03_State Management ii: World State Trie and Storage Trie](CN/03_state_management_stateTrie.md)
- [04_Transaction: 一个Transaction的生老病死](CN/04_transaction.md)
- [05_从Block到Blockchain: 链状区块结构的构建](CN/05_block_blockchain.md)
- [06_一个网吧老板是怎么用闲置的电脑进行挖矿的]
- [07_一个新节点是怎么加入网络并同步区块的]

### PART TWO - General Source Code Analysis: Lower-level Services

- [10_状态数据优化: Batch and Pruning]
- [11_Blockchain的数据是如何持久化的: Leveldb in Practice]
- [12_当I/O变成瓶颈: Snapshot in Practice]
- [13_深入EVM: 设计与实现]
- [14_Signer: 如何证明Transaction的合法性]
- [15_节点的调用 RPC and IPC](CN/15_rpc_ipc.md)

### PART THREE - Advanced Topics~P

- [20_结合BFT Consensus 解决拜占庭将军问题]
- [21_从Plasma到Rollup](CN/21_rollup.md)
- [22_Authenticated data structures Brief](CN/22_ads.md)
- [23_Bloom Filter](CN/23_bloom_filter.md)
- [24_图灵机和停机问题]
- [25_Log-structured merge-tree in Ethereum](CN/25_lsm_tree.md)
- [26_Concurrency in Ethereum Transaction](CN/26_txn_concurrency.md)
- [27_Zero-knowledge Proof]

### PART FOUR - Ethereum in Practice

- [30_使用geth构建一个私有网络](CN/30_geth_private_network.md)
- [31_如何编写Solidity语言](CN/31_solidity_in_practice.md)
- [32_使用预言机(Oracle)构建随机化的DApp]
- [33_Query On Ethereum Data]
- [34_layer2 in Practice]

### PART FIVE - APPENDIX

- [40_FQA](#tips)
- [41_Ethereum System Tunning]
- [42_go-ethereum的开发思想]
- [43_Metrics in Ethereum]
- [44_Golang & Ethereum]
- [45_Looking Forward to the Ethereum2.0]

-----------------------------------------------------------

## How to measure the level of understanding of a system？

如何衡量对一个系统的理解程度?

- Level 4: 掌握（Mastering）
  - 在完全理解的基础上，可以设计并编写一个新的系统
- Level 3: 完全理解（Complete Understanding）
  - 在理解的基础上，完全掌握系统的各项实现的细节，并能做出优化
  - 可以对现有的系统定制化到不同的应用场景
- Level 2: 理解（Understanding）
  - 熟练使用系统提供的API
  - 了解系统模块的调用关系
  - 能对系统的部分模块进行简单修改/重构
- Level 1:了解（Brief understanding）
  - 了解系统设计的目标，了解系统的应用场景
  - 可以使用系统的部分的API

 我们希望读者在阅读完本作之后，对Ethereum的理解能够达到 Level 2 - Level 3的水平。

## Some Details

- 以太坊是基于State状态机模型的区块链系统，Miner在update new Block的时候，会直接修改自身的状态（添加区块奖励给自己）。所以与Bitcoin不同的是，Ethereum的区块中，并没有类似的Coinbase的transaction。
- 在core/transaction.go 中, transaction的的数据结构是有time.Time的参数的。但是在下面的newTransaction的function中只是使用Local的time.now()对Transaction.time进行初始化。
- 在core/transaction.go 的transaction 数据结构定义的时候, 在transaction.time 后面的注释写到（Time first seen locally (spam avoidance), Time 只是用于在本地首次看到的时间。
- uncle block中的transaction 不会被包括到主链上。
- go-ethereum有专用函数来控制每次transaction执行完，返还给用户的Gas的量。有根据EIP-3529，每次最多返还50%的gas.
- 不同的Contracts的数据会混合的保存在底层的一个LevelDB instance中。
- LevelDB中保存的是MPT的Node信息。
- 在以太坊中，通常先执行Finalise函数 -> 然后执行Commit函数。

## 关键函数

```go
 // 向leveldb中更新Storage 数据
 func WritePreimages(db ethdb.KeyValueWriter, preimages map[common.Hash][]byte)

 // 向Blockchain中添加新的Block，会涉及到StateDB(Memory)/Trie(Memory)/EthDB(Disk)的更新
 func (bc *BlockChain) insertChain(chain types.Blocks, verifySeals, setHead bool) (int, error)

 // insertChain中调用来执行Block中的所有的交易
 func (p *StateProcessor) Process(block *types.Block, statedb *state.StateDB, cfg vm.Config) (types.Receipts, []*types.Log, uint64, error)

 //执行单条Transaction的调用
 func applyTransaction(msg types.Message, config *params.ChainConfig, bc ChainContext, author *common.Address, gp *GasPool, statedb *state.StateDB, blockNumber *big.Int, blockHash common.Hash, tx *types.Transaction, usedGas *uint64, evm *vm.EVM) (*types.Receipt, error)

 // 状态转移函数
 func (st *StateTransition) TransitionDb() (*ExecutionResult, error)

 // 执行合约内function
 func (in *EVMInterpreter) Run(contract *Contract, input []byte, readOnly bool) (ret []byte, err error)

 // opSstore的调用
 func (s *StateDB) SetState(addr common.Address, key, value common.Hash)
 // 被修改的state的值会首先被放在StateObject的dirtyStorage中，而不是直接添加到Trie或者Disk Database中。
 func (s *stateObject) setState(key, value common.Hash)

 // 在Finalizes所有的pending的Storage时候，并且更新到Trie，计算State Trie的Root
 func (s *StateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash

 // Finalise 当前内存中的Cache.
 func (s *StateDB) Finalise(deleteEmptyObjects bool) 

 // Commit StateDB中的Cache到内存数据库中
 func (s *StateDB) Commit(deleteEmptyObjects bool) (common.Hash, error)

 // 将StateObject中所有的dirtyStorage转存到PendingStorage中，并清空dirtyStorage，并给prefetcher赋值
 func (s *stateObject) finalise(prefetch bool)

 // 更新StorageObject对应的Trie, from Pending Storage
 func (s *stateObject) updateTrie(db Database) Trie

 // 最终获取到新的StateObject的Storage Root
 func (t *Trie) hashRoot() (node, node, error)

 // 用于在内存数据库中保存MPT节点
 func (c *committer) store(n node, db *Database) node

 // 向rawdb对应的数据库写数据(leveldb)
 func (db *Database) Commit(node common.Hash, report bool, callback func(common.Hash)) error

```

## References

本书主要参考了Go-Ethereum代码库，Ethereum Yellow Paper，以及EIP的具体Spec。读者可以从下面的Link中找到相关的引用资料。

- [1] Ethereum Yellow Paper [(Paper Link)](https://ethereum.github.io/yellowpaper/paper.pdf)
- [2] Ethereum/Go-Ethereum [(link)](https://github.com/ethereum/go-ethereum)
- [3] Go-ethereum code analysis [(link)](https://github.com/ZtesoftCS/go-ethereum-code-analysis)
- [4] Ethereum Improvement Proposals [(link)](https://github.com/ethereum/EIPs)
- [5] Mastering Bitcoin(Second Edition)
- [6] Mastering Ethereum [(link)](https://github.com/ethereumbook/ethereumbook)
- [7] EIP-2718: Typed Transaction Envelope[(link)](https://eips.ethereum.org/EIPS/eip-2718)
- [8] EIP-2930: Optional access lists [(link)](https://eips.ethereum.org/EIPS/eip-2930)
- [9] EIP-1559: Fee market change for ETH 1.0 chain [(link)](https://eips.ethereum.org/EIPS/eip-1559)
- [10] Ask about Geth: Snapshot acceleration [(link)](https://blog.ethereum.org/2020/07/17/ask-about-geth-snapshot-acceleration/)

## Talks

- Succinct Proofs in Ethereum - Barry Whitehat, Ethereum Foundation [(Youtube)](https://www.youtube.com/watch?v=TtsDNneTDDY)
