# StateDB & Trie & Secure Trie

## General

在本文中，我们剖析一下Ethereum State 管理模块中最重要的几个数据结构，StateDB, Trie，Secure Trie，以及StackTrie。我们讲通过分析Ethereum中的主workflow的方式来深入理解这三个数据结构的使用场景，以及设计上的不同。

首先，StateDB是这三个数据结构中最高层的封装，它是直接提供了与StateObject (Account，Contract)相关的CURD的接口给其他的模块，比如：

- Mining 模块，执行新Blockchain中的交易形成新的world state。
- Block同步模块，执行新Blockchain中的交易形成新的world state，与header中的state root进行比较验证。
- EVM中的两个与Contract中的持久化存储相关的两个opcode, sStore, sSload.

Trie结构是Ethereum中用于管理数据的基本数据结构，它被广泛的运用在Ethereum里的多个模块中，包括管理全局的World State Trie，管理Contract中持久化存储Key-Value 对的Storage Trie，以及每个Block中的Transaction Trie 和 Receipt Trie。

广义上的Trie的指的是Merkel Patricia Trie(MPT)，根据业务功能的不同，在go-ethereum中一共实现了三种不同的MPT，分别是，Trie，Secure Trie以及Stack Trie.

<!-- 这些Trie在具体实现上的不同点在于，Transaction Trie本质上并没有使用Trie来管理Transaction的数据，而是依赖于MPT的根来快速验证，具体可以参考core/types/hashing.go/DeriveSha()函数来了解Transaction Trie 的root是如何产生的，这里的Trie使用的是StackTrie。在本文中，我们主要研究的对象是与全局World State Trie有关的结构。 -->

从调用关系上看`Trie`是最底层的核心结构，它用于之间负责StateObject数据的保存，以及提供相应的CURD函数。它的定义在trie/trie.go文件中。

Secure Trie结构本质上是对Trie的一层封装。它具体的CURD操作的实现都是通过Trie中定义的函数来执行的。它的定义在trie/secure_trie.go文件中。目前StateDB中的直接对应的Trie是Secure Trie。这个Trie也就是我们常说的World State Trie，它是唯一的一个全局Trie。与Trie不同的是，Secure Trie要求新加入的Key-Value pair中的Key的数据都是Sha过的。这是为了方式恶意的构造Key来增加MPT的高度。

```go
type SecureTrie struct {
  trie             Trie
  hashKeyBuf       [common.HashLength]byte
  secKeyCache      map[string][]byte
  secKeyCacheOwner *SecureTrie // Pointer to self, replace the key cache on mismatch
}
```

不管是Secure Trie还是Trie，他们的创建的前提是更下层的db的实例已经创建成功了，否则就会报错。

值得注意的是一个关键函数Prove的实现并不在这两个Trie的定义文件中，而是位于trie/proof.go文件中。

## StateDB

我们可以在genesis block创建的相关代码中，找到直接相关的例子。

```go
 statedb.Commit(false)
 statedb.Database().TrieDB().Commit(root, true, nil)
```

具体World State的更新顺序是: statedb --> Memory Database (Memory State Trie) --> Disk (Leveldb Batch)

## Trie Operations

### Read Operation

### Insert

```go
func (t *Trie) insert(n node, prefix, key []byte, value node) (bool, node, error) {
 fmt.Println("Out n:", &n)
 if len(key) == 0 {
  if v, ok := n.(valueNode); ok {
   return !bytes.Equal(v, value.(valueNode)), value, nil
  }
  return true, value, nil
 }
 switch n := n.(type) {
 case *shortNode:
  matchlen := prefixLen(key, n.Key)
  // If the whole key matches, keep this short node as is
  if matchlen == len(n.Key) {
   dirty, nn, err := t.insert(n.Val, append(prefix, key[:matchlen]...), key[matchlen:], value)
   if !dirty || err != nil {
    return false, n, err
   }
   return true, &shortNode{n.Key, nn, t.newFlag()}, nil
  }
  // Otherwise branch out at the index where they differ.
  branch := &fullNode{flags: t.newFlag()}
  var err error
  _, branch.Children[n.Key[matchlen]], err = t.insert(nil, append(prefix, n.Key[:matchlen+1]...), n.Key[matchlen+1:], n.Val)
  if err != nil {
   return false, nil, err
  }
  _, branch.Children[key[matchlen]], err = t.insert(nil, append(prefix, key[:matchlen+1]...), key[matchlen+1:], value)
  if err != nil {
   return false, nil, err
  }
  // Replace this shortNode with the branch if it occurs at index 0.
  if matchlen == 0 {
   return true, branch, nil
  }
  // Otherwise, replace it with a short node leading up to the branch.
  return true, &shortNode{key[:matchlen], branch, t.newFlag()}, nil

 case *fullNode:
  dirty, nn, err := t.insert(n.Children[key[0]], append(prefix, key[0]), key[1:], value)
  if !dirty || err != nil {
   return false, n, err
  }
  n = n.copy()
  n.flags = t.newFlag()
  n.Children[key[0]] = nn
  return true, n, nil

 case nil:
  return true, &shortNode{key, value, t.newFlag()}, nil

 case hashNode:
  // We've hit a part of the trie that isn't loaded yet. Load
  // the node and insert into it. This leaves all child nodes on
  // the path to the value in the trie.
  rn, err := t.resolveHash(n, prefix)
  if err != nil {
   return false, nil, err
  }
  dirty, nn, err := t.insert(rn, prefix, key, value)
  if !dirty || err != nil {
   return false, rn, err
  }
  return true, nn, nil

 default:
  panic(fmt.Sprintf("%T: invalid node: %v", n, n))
 }
}
```

这里有一个关于go语言的知识。我们可以观察到insert函数的第一个参数是一个node类型的变量，变量名为n。有趣的是，在switch语句中我们看到了一个这样的写法.

```go
switch n := n.(type)
```

这种写法是合法的。

### Update

### Delete

### Finalize And Commit and Commit to Disk

- 在leveldb中保存的是Trie中的节点。
- <hash, node.rlprawdata>

## StackTrie

StackTrie虽然也是MPT结构，但是它与另外的两个Trie最大的不同在于，其主要作用不是用于存储数据，而是用于给一组数据生成证明。比如，在Block中的Transaction Hash以及Receipt Hash都是基于StackTrie生成的。这里我们使用一个更直观的例子。这个部分的代码位于*core/block_validator.go*中。在block_validator中定义了一系列验证用的函数, 比如`ValidateBody`和 `ValidateState`函数。我们选取了这两个函数的其中一部分，如下所示。为了验证Block的合法性，ValidateBody和ValidateState函数分别在本地基于Block中提供的数据来构造Transaction和Receipt的哈希来与Header中的TxHash与ReceiptHash。我们可以发现，函数`types.DeriveSha`需要一个`TrieHasher`类型的参数。但是在具体调用的时候，却传入了了一个`trie.NewStackTrie`类型的变量。这是因为StackTrie实现了TrieHasher接口所需要的三个函数，所以这种调用是合法的。我们可以在*core/types/hashing.go*中找到TrieHasher的定义。这里`DeriveSha`不断的向StackTrie中添加数据，并最终返回StackTrie的Root哈希值。

同时，我们可以发现，在调用DeriveSha函数的时候，我们每次都会new一个新的StackTrie作为参数。这也反映出了，StackTrie的主要作用就是生成验证用的Proof，而不是存储数据。

```golang
func (v *BlockValidator) ValidateBody(block *types.Block) error {
 ...
 if hash := types.DeriveSha(block.Transactions(), trie.NewStackTrie(nil)); hash != header.TxHash {
  return fmt.Errorf("transaction root hash mismatch: have %x, want %x", hash, header.TxHash)
 }
 ...
}
```

```golang
func (v *BlockValidator) ValidateState(block *types.Block, statedb *state.StateDB, receipts types.Receipts, usedGas uint64) error {
 ...
 // Tre receipt Trie's root (R = (Tr [[H1, R1], ... [Hn, Rn]]))
 receiptSha := types.DeriveSha(receipts, trie.NewStackTrie(nil))
 if receiptSha != header.ReceiptHash {
  return fmt.Errorf("invalid receipt root hash (remote: %x local: %x)", header.ReceiptHash, receiptSha)
 }
 ...
}

```

## Reference

- [1] <http://yangzhe.me/2019/01/18/ethereum-trie-part-2/>
