# Transaction Pool

## General 

交易可以分为 Local Transaction 和 Remote Transaction 两种。通过节点提供的 RPC 传入的交易，被划分为 Local Transaction，通过 P2P 网络传给节点的交易被划分为 Remote Transaction。

## 交易池的组成

Transaction Pool 主要有两个内存组件，Pending 和 Queue 组成。

## 交易池的限制

交易池设置了一些的参数来限制单个交易的 Size ，以及整个 Transaction Pool 中所保存的交易的总数量。当交易池的中维护的交易超过某个阈值的时候，交易池会丢弃/驱逐(Discard/Evict)一部分的交易。这里注意，被清除的交易通常都是 Remote Transaction，而 Local Transaction 通常都会被保留下来。

负责判断哪些交易会被丢弃的函数是 `txPricedList.Discard()`。

Transaction Pool 引入了一个名为 `txSlotSize` 的 Metrics 作为计量一个交易在交易池中所占的空间。目前，`txSlotSize` 的值是 `32 * 1024`。每个交易至少占据一个 `txSlot`，最大能占用四个 `txSlotSize`，`txMaxSize = 4 * txSlotSize = 128 KB`。换句话说，如果一个交易的物理数据大小不足 32KB，那么它会占据一个 `txSlot`。同时，一个合法的交易最大是 128KB 的大小

按照默认的设置，交易池的最多保存 `4096+1024` 个交易，其中 Pending 区保存 4096 个 `txSlot` 规模的交易，Queue 区最多保存 1024 个 `txSlot` 规模的交易。

## 