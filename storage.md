# 以太坊中的数据存储

## LevelDB

go-ethereum中使用leveldb作为本地存储的用的数据库。Leveldb是由Jeff Dean等人基于google Bigtable开发的单机模式下的key-value数据库，不支持sql语句。

在以太坊中，key通常与hash有关，value多为数据结构的RLP编码。

例如，以太坊中的区块头的存储结构为：

    rawdb/scheme.go: 
        // HeaderKey 由区块头前缀 + 区块号 + 区块hash组合
        headerHashKey = headerPrefix + num (uint64 big endian) + hash -> header
        // blockBodyKey 由区块体前缀 + 区块号 + 区块hash组合
        blockBodyKey = blockBodyPrefix + num (uint64 big endian) + hash -> block body

其中，Key中高位都是特定的前缀标示。
    rawdb/scheme.go:
        headerPrefix       = []byte("h") // headerPrefix + num (uint64 big endian) + hash -> header
        headerTDSuffix     = []byte("t") // headerPrefix + num (uint64 big endian) + hash + headerTDSuffix -> td
        headerHashSuffix   = []byte("n") // headerPrefix + num (uint64 big endian) + headerHashSuffix -> hash
        headerNumberPrefix = []byte("H") // headerNumberPrefix + hash -> num (uint64 big endian)

        blockBodyPrefix     = []byte("b") // blockBodyPrefix + num (uint64 big endian) + hash -> block body
        blockReceiptsPrefix = []byte("r") // blockReceiptsPrefix + num (uint64 big endian) + hash -> block receipts

        txLookupPrefix        = []byte("l") // txLookupPrefix + hash -> transaction/receipt lookup metadata
        bloomBitsPrefix       = []byte("B") // bloomBitsPrefix + bit (uint16 big endian) + section (uint64 big endian) + hash -> bloom bits
        SnapshotAccountPrefix = []byte("a") // SnapshotAccountPrefix + account hash -> account trie value
        SnapshotStoragePrefix = []byte("o") // SnapshotStoragePrefix + account hash + storage hash -> storage trie value
        codePrefix            = []byte("c") // codePrefix + code hash -> account code

## StateDB
