# 使用geth构建一个私有区块链网络 (Private Ethereum Blockchain)

## 创建 Private 网络

### 共识算法的选择

Geth支持多种共识算法，包括基于PoA(Proof-of-Authority)的Clique协议，以及PoW的Ethash协议。Clique协议在生成区块的时候没有Hash计算的压力，非常适合用于搭建测试链。比如，Ethereum两个著名的测试网络Rinkeby and Görli都是基于POA的共识算法搭建的。

值得注意的是，在使用Clique协议作为共识算法的网络中，mining Block是没有收益的。所以，如果你想讲Mining Reward作为你网络中的一项特性的话，请使用的Ethash协议。

### 构建创始区块

首先创建一个包含创世state信息的genesis.json文件，如下所示。注意，本例中的Genesis文件没有设置共识算法，所以Geth会默认使用Ethash协议作为共识算法。

在genesis.json中，config字段决定了整个链的一些基本的设定，通常这些设定在创世区块初始化之后之后是不可修改的。我们来介绍一下其中比较重要的一些设置。

首先是chainId字段。我们知道目前在市面上除了Ethereum主网之外，还有很多使用直接使用Geth，或者基于定制化修改的Geth来运行的公有/私有网络。这些客制化区块链网络非常之多，涉及到各个层面，比如Ethereum官方运营的两个测试链网络，币安运行的Binance Smart Chain，以及一些Layer-2的节点，比如Optimism。正如我们之前提到的，考虑到Geth节点之间的通信都是依赖于P2P网络传输，假如两个Geth-base的网络使用了相同的创始数据进行初始化，那么势必会造成混乱，网络数据同步带来混乱。因此，Ethereum的开发人员设计了chainId就是来解决这个问题。chainId用于在P2P的网络世界中来区分这些基于不同的版本/创世节点的网络信息，起到了网络身份证的作用。比如Ethereum主网的chainId是1，Binance Smart Chain主网的chainId是56。关于不同chainId对应的网络信息，可以参考这个[网站](https://chainlist.org/)。

alloc字段用于给一些账户初始化一些本网络的Native Token。在创世区块生成之后，网络中的Native Token产生的来源只有Mining。

```json
{
    "config": {
        "chainId": "Your Private Network id (Int number)",
        "homesteadBlock": 0,
        "eip150Block": 0,
        "eip155Block": 0,
        "eip158Block": 0,
        "byzantiumBlock": 0,
        "constantinopleBlock": 0,
        "petersburgBlock": 0,
        "istanbulBlock": 0,
        "berlinBlock": 0
    },
    "alloc": {
        "Your Account Address": {
            "balance": "10000000000000000000"
        }
    },
    "coinbase": "0x0000000000000000000000000000000000000000",
    "difficulty": "0x200",
    "extraData": "",
    "gasLimit": "0x2fefd8",
    "nonce": "0x00000000000000755",
    "mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "timestamp": "0x00"
}
```

### 构建网络

#### 初始化创世区块

在每个网络节点运行之前，都需要基于创始节点进行初始化。命令是init 加上 --datadir 后面跟上需要存放Chain数据文件路径，最后跟上链的创世区块genesis.json文件。

```cmd
geth init --datadir  <Datadir> genesis.json
```

#### 运行节点

```cmd
geth --datadir  <Datadir>  --networkid <networkid> --nodiscover --http --rpc --rpcport "8545" --rpcaddr "0.0.0.0" --rpccorsdomain "*" --rpcapi "eth,web3,net,personal,miner" console 2
```

## Monitoring

区块链浏览器可以方便的查询链上数据。一般来说，它们通过go-ethereum实例的RPC接口来调用，实例的API，从而获取最新的链上信息。