# 以太坊中的receipt

receipt是Ethereum基础数据结构的一种，源代码位于core/types/receipt.go。

用源代码话中来概括receipt是：receipt用于表示transaction的结果。

Receipt的具体数据结构如下所示。

    PostState []byte
    Status uint64
    CumulativeGasUsed uint64
    Bloom Bloom
    Logs []*Log

    TxHash common.Hash
    ContractAddress common.Address
    GasUsed uint64

    BlockHash common.Hash
    BlockNumber *big.Int
    TransactionIndex uint
