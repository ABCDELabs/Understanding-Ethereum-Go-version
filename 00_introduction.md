### go-ethereum目录解析
go-ethereum项目进行过若干次的重构，本文基于最新的版本Marljeh (version-1.9.25 updated time 2020-12) 进行分析。

目前，go-ethereum项目的目录结构如下所示。

	accounts/       	实现了一个高等级的以太坊账户管理
	build/			主要是编译和构建的一些脚本
	accounts/
	 ├──abi			解析Contracts中的ABI的信息
	 	├──abi.go	
	core/			以太坊核心模块，包括核心数据结构，状态树及其算法实现
	 ├──state/
	 ├──types/		包括Block在内的以太坊核心数据结构
	 	├──block.go		以太坊block
		├──bloom9.go		一个Bloom Filter的实现
		├──transaction.go	以太坊transaction的数据结构与实现
		|──transaction_signing.go	用于对transaction进行签名的函数的实现
		|──tx_pool.go
		├──receipt.go		以太坊收据的实现，用于说明以太坊交易的结果
	├──miner/
		├──miner.go			矿工的基本的实现。
		├──worker.go		矿工任务的模块，包括打包transaction
		├──unconfirmed.go
	├──consensus/
		├──consensus.go		共识相关的参数设定，包括Block Reward的数量
	├──state/
		├──statedb.go		StateDB结构用于存储所有的与Merkle trie相关的存储, 包括一些循环state结构
	├──trie/				package trie包含了Merkle Patricia Tries的实现
		├──trie.go
	├──rlp/					RLP的Encode与Decode的相关实现
	├──ethdb/				Ethereum 本地存储的相关实现, 包括leveldb的调用
		├──leveldb/			Go-Ethereum使用的与Bitcoin Core version一样的Leveldb作为本机存储用的数据库
		├──database.go		Database 服务的各种接口
	├──node/				
	├──rpc/					Ethereum RPC客户端的实现
	├──p2p/					Ethereum 使用的P2P网络的实现,包括节点发现，节点链接等
	├──les/					Ethereum light client的实现
