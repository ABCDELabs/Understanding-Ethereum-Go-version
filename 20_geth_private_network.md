# 使用geth构建一个私有区块链网络 (Private Ethereum Blockchain)


## 创建 Private 网络
首先创建一个包含创世state信息的genesis.json文件，如下所示。

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
    "difficulty": "0x20000",
    "extraData": "",
    "gasLimit": "0x2fefd8",
    "nonce": "0x00000000000000755",
    "mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "timestamp": "0x00"
}
```

使用初始化命令：
```cmd
geth init --datadir  <Datadir> genesis.json
```

运行节点：
```cmd
geth --datadir  <Datadir>  --networkid <networkid> --nodiscover --http --rpc --rpcport "8545" --rpcaddr "0.0.0.0" --rpccorsdomain "*" --rpcapi "eth,web3,net,personal,miner" console 2
```

## 使用区块链浏览器来查询链上数据

区块链浏览器可以方便的查询链上数据。一般来说，它们通过go-ethereum实例的RPC接口来调用，实例的API，从而获取最新的链上信息。