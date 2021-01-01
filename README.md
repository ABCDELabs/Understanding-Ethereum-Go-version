# Understanding-Ethereum-Go-version
Understanding Ethereum(Go version)｜理解以太坊(Go 版本源码剖析)

Author: Siyuan Han 


本文档基于Go-Ethereum (Marljeh version-1.9.25)对以太坊的源码结构，以及以太坊系统设计背后的细节，原理进行剖析。

go-ethereum是以太坊协议的Go语言实现版本。除了本版本之外，Ethereum还有C++, Python，Java等其他语言版本。Go-ethereum在这些所有的社区版本中，版本更新最频繁，开发人员最多，问题相对较少。其他语言的Ethereum实现版本因为，更新频率相对较低，隐藏问题未知，建议初学者首先从go-ethereum的视角来理解Ethereum网络与系统的设计实现。

### go-ethereum目录解析
go-ethereum项目进行过若干次的重构，本文基于最新的版本Marljeh (version-1.9.25 updated time 2020-12) 进行分析。

目前，go-ethereum项目的目录结构如下所示。

	accounts        	实现了一个高等级的以太坊账户管理
	build			主要是编译和构建的一些脚本
	core			以太坊核心模块，包括核心数据结构，状态树及其算法实现
	├──types.go		包括Block在内的以太坊核心数据结构


## Reference
-  Go-ethereum code analysis https://github.com/ZtesoftCS/go-ethereum-code-analysis
- Mastering Bitcoin(Second Edition)
