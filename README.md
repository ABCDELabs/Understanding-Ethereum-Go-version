# Understanding-Ethereum-Go-version

- Title: Understanding Ethereum(Go version)｜理解以太坊(Go 版本源码剖析)
- Subject: Ethereum Source Code Analysis | 以太坊源码分析
- Author: Siyuan Han
- Updated date: 2021-07

## Preface

Blockchain作为过去几年技术社区最热点话题之一, 每当我们提到它的时候，首先就会讨论到成功运用这项技术的最火热的几个系统。但是不管是讨论到以加密货币导向（Crypto-based）的Bitcoin Network, 还是致力于实现通用框架（General-Purpose）的Ethereum的时候，通常的文档往往只是在high-level的层面来讲述他们的架构。现在的技术社区有非常多的文档来讲述，这些Blockchain System背后的数据结构，以及类似双花，梅克尔树等区块链系统的专有问题。但是某天，我忽然想到，究竟miner是怎么从transaction pool中选取transaction，他们又是按照怎么的order被打包进区块链中的呢？我尝试去搜索了一下，发现鲜有文章提到这一层面的细节。本文作为我学习的记录，将会从源码的角度来深度解析区块链系统中各个模块的实现的细节。

笔者坚信，在未来的是五到十年内，这个世界的云端服务一定是两极分化的。一极是以大云计算公司（ie： Google，MS，Oracle，Snowflake，Alibaba）为代表的中心化服务，另一极就是以Blockchain技术作为核心的去中心化的世界。在这个世界中，Ethereum是当之无愧的领头羊。Ethereum 不光在Public Chain的层面取得了巨大的成功，而且Go-Ehtereum作为其优秀的开源实现，已经被广泛的订制，来适应不同的私有/联盟场景。所以，要想真正掌握好区块链系统的实现，研究好Ethereum的原理以及其设计思想是非常有必要。
 区版本中，版本更新最频繁，开发人员最多，问题相对较少。其他语言的Ethereum实现版本因为更新频率相对较低，隐藏问题未知，建议初学者首先从go-ethereum的视角来理解Ethereum网络与系统的设计实现。

## Contents

### PART ONE - General Source Code Analysis

- [00_万物的起点从geth出发: Geth框架导引](00_geth.md) 
- [01_State-based 模型 & Account](01_account.md) 
- [02_Transaction是怎么被打包的: 一个Transaction的生老病死](02_transaction.md)
- [03_从Block到Blockchain: 区块链数据结构的构建]() 
- [04_一个新节点是怎么加入网络并同步区块的]() 
- [05_一个网吧老板是怎么用闲置的电脑进行挖矿的]() 

### PART TWO - How to use go-ethereum

- [10_构建一个StateDB的实例]()

### PART THREE - Ethereum in Practice

- [20_使用geth构建一个私有网络](20_geth_private_network.md)
- [21_如何编写Solidity语言]()
- [22_使用预言机(Oracle)构建随机化的DApp]()
- [23_Query On Ethereum Data]()


### PART FOUR - Advanced Topics

- [30_结合BFT Consensus 解决拜占庭将军问题]()
- [31_Plasma与 Zk Rollup]()
- [32_ADS]()
- [33_Bloom Filter]()
- [34_图灵机和停机问题]()
- [35_Log-structured merge-tree in Ethereum]()
- [36_Ethereum Transaction Currency]()

### PART FIVE - APPENDIX

- [40_FQA](#tips)
- [41_Ethereum System Tunning]()
- [42_go-ethereum的开发思想](42_developer_view.md)
- [43_Metrics in Ethereum](43_metrics.md)

## 为什么要阅读区块链系统的源代码

1. 文档资料相对较少，且**内容浅尝辄止**。比如，很多的科普文章都提到，miner负责把transactions从transaction pool中打包到新的block中。那么：
    - miner是怎么从transaction pool中选取这些transaction的呢？
    - 被选择的transaction又是以怎样的顺序(Order)被打包到区块中的呢？
    - 在执行transaction的EVM是怎么计算gas used?
    - 剩余的gas又是怎么返还给Transaction Proposer的呢？
    - 在执行transaction中是哪个模块，又是怎样去修改Contract中的持久化变量呢？
    - Contract中的持久化变量又是以什么样的形式存储的呢？

2. 目前的Blockchain系统并没有像数据库系统(DBMS)那样统一实现的方法论，每个不同的系统中都集成了大量的细节。如果不从源码的角度入手，很多的细节容易被忽略掉。简单的说，一个完整的区块链系统至少包含以下的模块: 
    - 密码学模块: 加解密，签名，安全hash，mining
    - 网络模块: P2P节点通信
    - 分布式共识模块: PoW, BFT
    - 智能合约解释器模块: Solidity编译语言，EVM解释器
    - 数据存储模块: 数据库，数据存储，index

-----------------------------------------------------------

## Conclusion

如何衡量对一个系统的理解程度

1. 掌握（Mastering）
    - 可以编写一个新的系统
2. 完全理解（Complete Understanding）
    - 完全理解系统的各项实现的细节，并能做出优化
    - 可以对现有的系统定制化到不同的应用场景
3. 理解（Understanding）
    - 熟练使用系统提供的API
    - 能对系统的部分模块进行重构
4. 简单了解（Brief understanding）
    - 了解系统设计的目标，了解系统的应用场景
    - 可以使用系统的部分的API

## Tips
<a name="tips"></a>

- 以太坊是基于State模型的区块链系统，miner在update new Block的时候，会直接修改自身的状态（添加区块奖励给自己）。所以与Bitcoin不同的是，Ethereum的区块中，并没有类似的Coinbase的transaction。
- 在core/transaction.go 中, transaction的的数据结构是有time.Time的参数的。但是在下面的newTransaction的function中只是使用Local的time.now()对Transaction.time进行初始化。
- 在core/transaction.go 的transaction 数据结构定义的时候, 在transaction.time 后面的注释写到（Time first seen locally (spam avoidance)）, Time 只是用于在本地首次看到的时间。
- uncle block中的transaction 不会被包括到主链上。
- go-ethereum有专用函数来控制每次transaction执行完，返还给用户的Gas的量。有根据EIP-3529，每次最多返还50%的gas.

## Reference

- [1] Etheruem Yellow Paper [(Paper Link)](https://ethereum.github.io/yellowpaper/paper.pdf)
- [2] Ethereum/Go-Ethereum [(link)](https://github.com/ethereum/go-ethereum)
- [3] Go-ethereum code analysis [(Link)](https://github.com/ZtesoftCS/go-ethereum-code-analysis) 
- [4] Ethereum Improvement Proposals [(link)](https://github.com/ethereum/EIPs)
- [5] Mastering Bitcoin(Second Edition)
- [6] Mastering Ethereum [(link)](https://github.com/ethereumbook/ethereumbook)

## Talks

- Succinct Proofs in Ethereum - Barry Whitehat, Ethereum Foundation [(Youtube)](https://www.youtube.com/watch?v=TtsDNneTDDY)
