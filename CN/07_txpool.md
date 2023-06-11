# Transaction Pool

## 概述 

交易可以分为 Local Transaction 和 Remote Transaction 两种。通过节点提供的 RPC 传入的交易，被划分为 Local Transaction，通过 P2P 网络传给节点的交易被划分为 Remote Transaction。

## 交易池的基本结构

Transaction Pool 主要有两个内存组件，`Pending` 和 `Queue` 组成。具体的定义如下所示。

```go
type TxPool struct {
	config      Config
	chainconfig *params.ChainConfig
	chain       blockChain
	gasPrice    *big.Int
	txFeed      event.Feed
	scope       event.SubscriptionScope
	signer      types.Signer
	mu          sync.RWMutex

	istanbul bool // Fork indicator whether we are in the istanbul stage.
	eip2718  bool // Fork indicator whether we are using EIP-2718 type transactions.
	eip1559  bool // Fork indicator whether we are using EIP-1559 type transactions.
	shanghai bool // Fork indicator whether we are in the Shanghai stage.

	currentState  *state.StateDB // Current state in the blockchain head
	pendingNonces *noncer        // Pending state tracking virtual nonces
	currentMaxGas uint64         // Current gas limit for transaction caps

	locals  *accountSet // Set of local transaction to exempt from eviction rules
	journal *journal    // Journal of local transaction to back up to disk

	pending map[common.Address]*list     // All currently processable transactions
	queue   map[common.Address]*list     // Queued but non-processable transactions
	beats   map[common.Address]time.Time // Last heartbeat from each known account
	all     *lookup                      // All transactions to allow lookups
	priced  *pricedList                  // All transactions sorted by price

	chainHeadCh     chan core.ChainHeadEvent
	chainHeadSub    event.Subscription
	reqResetCh      chan *txpoolResetRequest
	reqPromoteCh    chan *accountSet
	queueTxEventCh  chan *types.Transaction
	reorgDoneCh     chan chan struct{}
	reorgShutdownCh chan struct{}  // requests shutdown of scheduleReorgLoop
	wg              sync.WaitGroup // tracks loop, scheduleReorgLoop
	initDoneCh      chan struct{}  // is closed once the pool is initialized (for tests)

	changesSinceReorg int // A counter for how many drops we've performed in-between reorg.
}

```

## 交易池的限制

交易池设置了一些的参数来限制单个交易的 Size ，以及整个 Transaction Pool 中所保存的交易的总数量。当交易池的中维护的交易超过某个阈值的时候，交易池会丢弃/驱逐(Discard/Evict)一部分的交易。这里注意，被清除的交易通常都是 Remote Transaction，而 Local Transaction 通常都会被保留下来。

负责判断哪些交易会被丢弃的函数是 `txPricedList.Discard()`。

Transaction Pool 引入了一个名为 `txSlotSize` 的 Metrics 作为计量一个交易在交易池中所占的空间。目前，`txSlotSize` 的值是 `32 * 1024`。每个交易至少占据一个 `txSlot`，最大能占用四个 `txSlotSize`，`txMaxSize = 4 * txSlotSize = 128 KB`。换句话说，如果一个交易的物理数据大小不足 32KB，那么它会占据一个 `txSlot`。同时，一个合法的交易最大是 128KB 的大小

按照默认的设置，交易池的最多保存 `4096+1024` 个交易，其中 Pending 区保存 4096 个 `txSlot` 规模的交易，Queue 区最多保存 1024 个 `txSlot` 规模的交易。

## 交易池的更新



