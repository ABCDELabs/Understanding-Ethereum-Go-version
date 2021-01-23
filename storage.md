# 以太坊中的数据存储

go-ethereum中使用leveldb作为本地存储的用的数据库。Leveldb是由Jeff Dean等人基于google Bigtable开发的单机模式下的key-value数据库，不支持sql语句。

在以太坊中，key通常与hash有关，value多为数据结构的RLP编码。

例如，以太坊中的区块头的存储结构为：

    rawdb/scheme.go: 
        // HeaderKey 由区块头前缀 + 区块号 + 区块hash组合
        headerHashKey = headerPrefix + num (uint64 big endian) + hash -> header
        // blockBodyKey 由区块体前缀 + 区块号 + 区块hash组合
        blockBodyKey = blockBodyPrefix + num (uint64 big endian) + hash -> block body

其中，Key中高位都是特定的前缀标示。


