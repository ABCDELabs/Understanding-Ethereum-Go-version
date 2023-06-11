# 区块和区块链 (Block & Blockchain)

## Block 

### 基础数据结构

```go
type Block struct {
 header       *Header
 uncles       []*Header
 transactions Transactions
 hash atomic.Value
 size atomic.Value
 td *big.Int
 ReceivedAt   time.Time
 ReceivedFrom interface{}
}
```

```go
type Header struct {
 ParentHash  common.Hash    `json:"parentHash"       gencodec:"required"`
 UncleHash   common.Hash    `json:"sha3Uncles"       gencodec:"required"`
 Coinbase    common.Address `json:"miner"            gencodec:"required"`
 Root        common.Hash    `json:"stateRoot"        gencodec:"required"`
 TxHash      common.Hash    `json:"transactionsRoot" gencodec:"required"`
 ReceiptHash common.Hash    `json:"receiptsRoot"     gencodec:"required"`
 Bloom       Bloom          `json:"logsBloom"        gencodec:"required"`
 Difficulty  *big.Int       `json:"difficulty"       gencodec:"required"`
 Number      *big.Int       `json:"number"           gencodec:"required"`
 GasLimit    uint64         `json:"gasLimit"         gencodec:"required"`
 GasUsed     uint64         `json:"gasUsed"          gencodec:"required"`
 Time        uint64         `json:"timestamp"        gencodec:"required"`
 Extra       []byte         `json:"extraData"        gencodec:"required"`
 MixDigest   common.Hash    `json:"mixHash"`
 Nonce       BlockNonce     `json:"nonce"`
 // BaseFee was added by EIP-1559 and is ignored in legacy headers.
 BaseFee *big.Int `json:"baseFeePerGas" rlp:"optional"`
}
```

## Blockchain：区块链

### 基础数据结构

```golang
type BlockChain struct {
 chainConfig *params.ChainConfig // Chain & network configuration
 cacheConfig *CacheConfig        // Cache configuration for pruning

 db     ethdb.Database // Low level persistent database to store final content in
 snaps  *snapshot.Tree // Snapshot tree for fast trie leaf access
 triegc *prque.Prque   // Priority queue mapping block numbers to tries to gc
 gcproc time.Duration  // Accumulates canonical block processing for trie dumping

 // txLookupLimit is the maximum number of blocks from head whose tx indices
 // are reserved:
 //  * 0:   means no limit and regenerate any missing indexes
 //  * N:   means N block limit [HEAD-N+1, HEAD] and delete extra indexes
 //  * nil: disable tx reindexer/deleter, but still index new blocks
 txLookupLimit uint64

 hc            *HeaderChain
 rmLogsFeed    event.Feed
 chainFeed     event.Feed
 chainSideFeed event.Feed
 chainHeadFeed event.Feed
 logsFeed      event.Feed
 blockProcFeed event.Feed
 scope         event.SubscriptionScope
 genesisBlock  *types.Block

 // This mutex synchronizes chain write operations.
 // Readers don't need to take it, they can just read the database.
 chainmu *syncx.ClosableMutex

 currentBlock     atomic.Value // Current head of the block chain
 currentFastBlock atomic.Value // Current head of the fast-sync chain (may be above the block chain!)

 stateCache    state.Database // State database to reuse between imports (contains state cache)
 bodyCache     *lru.Cache     // Cache for the most recent block bodies
 bodyRLPCache  *lru.Cache     // Cache for the most recent block bodies in RLP encoded format
 receiptsCache *lru.Cache     // Cache for the most recent receipts per block
 blockCache    *lru.Cache     // Cache for the most recent entire blocks
 txLookupCache *lru.Cache     // Cache for the most recent transaction lookup data.
 futureBlocks  *lru.Cache     // future blocks are blocks added for later processing

 wg            sync.WaitGroup //
 quit          chan struct{}  // shutdown signal, closed in Stop.
 running       int32          // 0 if chain is running, 1 when stopped
 procInterrupt int32          // interrupt signaler for block processing

 engine     consensus.Engine
 validator  Validator // Block and state validator interface
 prefetcher Prefetcher
 processor  Processor // Block transaction processor interface
 forker     *ForkChoice
 vmConfig   vm.Config
}

```