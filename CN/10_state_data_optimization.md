# State Data Optimization：状态数据的优化

在之前的文章中我们提到了，State Trie/Storage Trie中的数据的写盘是依赖于`trie/database.go`这个文件中提供的函数。这个文件同样起到了Trie结构优化的作用

1. Batch Write
2. Node Pruning
   1. 节点的引用计数
   2. 当节点的引用计数为0时，从数据库中删除掉这个节点。

## Batch Write

## State Pruning
