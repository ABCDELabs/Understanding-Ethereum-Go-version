# 状态管理 ii: State Trie and Storage Trie

写在前面: 在最新的 `geth` 代码库中，`SecureTrie` 已经被重命名为了 `StateTrie`，相关的代码功能也进行了些许调整。因此，为了避免歧义，我们在这里提醒读者 **StateTrie 就是之前的 SecureTrie**。读者在阅读其他的文档时，如果遇到了 `SecureTrie`, 可以将其理解为 `StateTrie`。

## 理解 Trie 结构

Trie 结构是 Ethereum 中用于管理数据的基本数据结构，它被广泛的运用在Ethereum 里的多个模块中，包括管理全局的 State Trie，管理 Contract中持久化存储的Storage Trie，以及每个 Block 中的与交易相关的 Transaction Trie 和 Receipt Trie。

在以太坊的体系中，广义上的 Trie 的指的是 Merkel Patricia Trie (MPT)这种数据结构。在实际的实现中，根据业务功能的不同，在 go-ethereum 中一共实现了三种不同的 MPT 的 instance，分别是，`Trie`，`State Trie`(`Secure Trie`) 以及`Stack Trie`。由于已经有大量的资料来介绍 MPT 的具体数据结构，在本文中我们就不对 MPT 具体的结构进行解析，感兴趣的读者可以自行搜索相应的资料。


从调用关系上看 `Trie` 是最底层的核心结构，它用于之间负责 StateObject 数据的保存，以及提供相应的 CURD 函数。它的定义在 `trie/trie.go` 文件中。

State Trie 结构本质上是对 Trie 的一层封装。它具体的CURD操作的实现都是通过`Trie`中定义的函数来执行的。它的定义在`trie/secure_trie.go`文件中。目前StateDB 中的使用的 Trie 是经过封装之后的 State Trie。这个 Trie 也就是我们常说的 State Trie，它是唯一的一个全局 Trie 结构。与 Trie 不同的是，Secure Trie 要求新加入的 Key-Value pair 中的 Key 的数据都是 Sha函数哈希过的。这是为了防止恶意的构造 Key 来增加 MPT 的高度。

```go
type StateTrie struct {
 trie             Trie
 preimages        *preimageStore
 hashKeyBuf       [common.HashLength]byte
 secKeyCache      map[string][]byte
 secKeyCacheOwner *StateTrie // Pointer to self, replace the key cache on mismatch
}
```

不管是 Secure Trie还是Trie，他们的创建的前提是: 更下层的 db 的实例已经创建成功了，否则就会报错。值得注意的是一个关键函数 Prove 的实现，并不在这两个Trie的定义文件中，而是位于`trie/proof.go`文件中。

值得注意的是一个关键函数Prove的实现，并不在这两个Trie的定义文件中，而是位于`trie/proof.go`文件中。

## Trie 运用

### Read Operation：读写行动

具体的读取 Trie 上节点的数据是通过 `tryGet()` 函数来实现的。

```go
func (t *Trie) tryGet(origNode node, key []byte, pos int) (value []byte, newnode node, didResolve bool, err error) {
	switch n := (origNode).(type) {
	case nil:
		return nil, nil, false, nil
	case valueNode:
		return n, n, false, nil
	case *shortNode:
		if len(key)-pos < len(n.Key) || !bytes.Equal(n.Key, key[pos:pos+len(n.Key)]) {
			// key not found in trie
			return nil, n, false, nil
		}
		value, newnode, didResolve, err = t.tryGet(n.Val, key, pos+len(n.Key))
		if err == nil && didResolve {
			n = n.copy()
			n.Val = newnode
		}
		return value, n, didResolve, err
	case *fullNode:
		value, newnode, didResolve, err = t.tryGet(n.Children[key[pos]], key, pos+1)
		if err == nil && didResolve {
			n = n.copy()
			n.Children[key[pos]] = newnode
		}
		return value, n, didResolve, err
	case hashNode:
		child, err := t.resolveAndTrack(n, key[:pos])
		if err != nil {
			return nil, n, true, err
		}
		value, newnode, _, err := t.tryGet(child, key, pos)
		return value, newnode, true, err
	default:
		panic(fmt.Sprintf("%T: invalid node: %v", origNode, origNode))
	}
}
```

### Insert
在 Trie 上插入新节点是通过 `insert()` 函数来实现的。

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

这里有一个关于go语言的知识：我们可以观察到insert函数的第一个参数是一个变量名为n的node类型的变量。有趣的是，在switch语句中我们看到了一个这样的写法：

```go
switch n := n.(type)
```

显然语句两端的*n*的含义并不相同。这种写法在go中是合法的。


### Finalize And Commit to Disk：存储到硬盘

在更底层的 leveldb中，KV保存的是Trie中的节点，<hash, node.rlprawdata>。在Geth中，Trie并不是实时更新的，而是依赖于 Committer 和 Database 两个额外的辅助组件。

```
Trie.Commit --> Committer.Commit --> trie/Database.insert
```

事实上，由于缓存机制， Trie 的Commit并不会真的对Disk Database的值进行修改。Trie 真正更新到 Disk Database 的，是依赖于 `trie/Database.Commit` 函数的调用。我们可以在诸多函数中找到这个函数的调用比如。

```go
func GenerateChain(config *params.ChainConfig, parent *types.Block, engine consensus.Engine, db ethdb.Database, n int, gen func(int, *BlockGen)) ([]*types.Block, []types.Receipts) {
  ...
   // Write state changes to db
   root, err := statedb.Commit(config.IsEIP158(b.header.Number))
   if err != nil {
    panic(fmt.Sprintf("state write error: %v", err))
   }
   if err := statedb.Database().TrieDB().Commit(root, false, nil); err != nil {
    panic(fmt.Sprintf("trie write error: %v", err))
   }
   ...
}
```

#### State Trie 的更新是什么时候发生的？
  
  State Trie的 更新，通常是指的是基于State Trie中节点值的变化从而重新计算State Trie的Root的Hash值的过程。目前这一过程是通过调用StateDB中的`IntermediateRoot`函数来完成的。
  
  我们从三个粒度层面来看待State Trie更新的问题。

- Block 层。
    在一个新的Block Insert到Blockchain的过程中，State Trie可能会发生多次的更新。比如，在每次Transaction被执行之后，`IntermediateRoot`函数都会被调用。同时，更新后的 State Trie的Root值，会被写入到Transaction对应的Receipt中。请注意，在调用`IntermediateRoot`函数时，更新后的值在此时并没有被立刻写入到Disk Database中。此时的State Trie Root只是基于内存中的数据计算出来的。真正的Trie数据写盘，需要等到`trieDB.Commit`函数的执行。
- Transaction 层。
    如上面提到的，在每次Transaction执行完成后，系统都会调用一次StateDB的`IntermediateRoot`函数，来更新State Trie。并且会将更新后的Trie的Root Hash写入到该Transaction对应的Receipt中。这里提一下关于`IntermediateRoot`函数细节。在IntermediateRoot`函数调用时，会首先更新被修改的Contract的Storage Trie的Root。
- Instruction 层。
    执行Contract的Instruction，并不会直接的引发State Trie的更新。比如，我们知道，EVM指令`OpSstore`会修改Contract中的持久化存储。这个指令调用了StateDB中的`SetState`函数，并最终调用了对应的StateObject中的`setState`函数。StateObject中的`setState` 函数并没有直接对Contract的Storage Trie进行更新，而是将修改的存储对象保存在了StateObject中的*dirtyStorage* 中(*dirtyStorage*是用于缓存Storage Slot数据的Key-Value Map). Storage Trie的更新是由更上层的函数调用所触发的，比如`IntermediateRoot`函数，以及`StateDB.Commit`函数。

## StackTrie

`StackTrie` 虽然也是 `MPT` 结构，但是相比于作为索引结构来管理数据，`StackTrie`更直接的一个用法是给一组数据生成证明。例如，在 `Block` 中的 `Transaction Hash` 以及` Receipt Hash` 都是基于 `StackTrie` 生成的。这里我们使用一个更直观的例子。这个部分的代码位于 `core/block_validator.go` 中。在 `block_validator` 中定义了一系列验证用的函数, 比如 `ValidateBody()` 和 `ValidateState()` 函数。我们选取了这两个函数的其中一部分，如下所示。为了验证 Block 的合法性，`ValidateBody()` 和 `ValidateState()` 函数分别在本地基于 Block 中提供的数据来构造 Transaction 和 Receipt 的哈希来与Header 中的 `TxHash` 与 `ReceiptHash`。我们可以发现，函数 `types.DeriveSha` 需要一个 `TrieHasher` 类型的参数。但是在具体调用的时候，却传入了了一个`trie.NewStackTrie` 类型的变量。这是因为 `StackTrie` 实现了 `TrieHasher` 接口所需要的三个函数，所以这种调用是合法的。我们可以在 `core/types/hashing.go` 中找到 `TrieHasher` 的定义。这里 `DeriveSha` 不断的向 `StackTrie` 中添加数据，并最终返回 `StackTrie` 的 Root 哈希值用作数据证明。

```go
func (v *BlockValidator) ValidateBody(block *types.Block) error {
 ...
 if hash := types.DeriveSha(block.Transactions(), trie.NewStackTrie(nil)); hash != header.TxHash {
  return fmt.Errorf("transaction root hash mismatch: have %x, want %x", hash, header.TxHash)
 }
 ...
}
```

```go
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

同时，我们可以发现，在调用 `DeriveSha` 函数的时候，我们每次都会 `new` 一个新的 `StackTrie` 作为参数。这也反映出了，在这个部分 `StackTrie` 的主要作用就是生成验证用的Proof，而不是像管理账户状态的 `State Trie`一样唯一存在。




