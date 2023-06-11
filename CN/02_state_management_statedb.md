# 状态管理一: StateDB

## 概述

在本章中，我们来简析一下 go-ethereum 状态管理模块 StateDB。

## 理解 StateDB 的结构

我们知道以太坊是是基于以账户为核心的状态机 (State Machine)的模型。在账户的值发生变化的时候，我们说该账户从一个状态转换到了另一个状态。我们知道，在实际中，每个地址都对应了一个账户。随着以太坊用户和合约数量的增加，如何管理这些账户是客户端开发人员需要解决的首要问题。在 go-ethereum 中，StateDB 模块就是为管理账户状态设计的。它是直接提供了与 `StateObject` (账户和合约的抽象) 相关的 CURD 的接口给其他的模块，比如：

- (这个模块在 Merge 之后被废弃) Mining 模块: 执行新 Block 中的交易时调用 `StateDB` 来更新对应账户的值，并且形成新 world state。
- Block 同步模块，执行新 Block 中的交易时调用 `StateDB` 来更新对应账户的值，并且形成新 world state，同时用这个计算出来的 world state与 Block Header 中提供的 state root 进行比较，来验证区块的合法性。
- 在 EVM 模块中，调用与合约存储有关的相关的两个 opcode, `sStore` 和 `sSload` 时会调用 `StateDB`中的函数来查询和更新 Contract 中的持久化存储.
- ...

在实际中，所有的账户数据(包括当前和历史的状态数据)最终还是持久化在硬盘中。目前所有的状态数据都是通过KV的形式被持久化到了基于 LSM-Tree 的存储引擎中(例如 go-leveldb)。显然，直接从这种KV存储引擎中读取和更新状态数据是不友好的。而 `StateDB` 就是为了操作这些数据而诞生的抽象层。`StateDB` 本质上是一个用于管理所有账户状态的位于内存中的抽象组件。从某种意义上说，我们可以把它理解成一个中间层的内存数据库。


`StateDB` 的定义位于 `core/state/statedb.go` 文件中，如下所示。

```go
type StateDB struct {
	db         Database
	prefetcher *triePrefetcher
	trie       Trie
	hasher     crypto.KeccakState

	// originalRoot is the pre-state root, before any changes were made.
	// It will be updated when the Commit is called.
	originalRoot common.Hash

	snaps        *snapshot.Tree
	snap         snapshot.Snapshot
	snapAccounts map[common.Hash][]byte
	snapStorage  map[common.Hash]map[common.Hash][]byte

	// This map holds 'live' objects, which will get modified while processing a state transition.
	stateObjects         map[common.Address]*stateObject
	stateObjectsPending  map[common.Address]struct{} // State objects finalized but not yet written to the trie
	stateObjectsDirty    map[common.Address]struct{} // State objects modified in the current execution
	stateObjectsDestruct map[common.Address]struct{} // State objects destructed in the block

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// The refund counter, also used by state transitioning.
	refund uint64

	thash   common.Hash
	txIndex int
	logs    map[common.Hash][]*types.Log
	logSize uint

	preimages map[common.Hash][]byte

	// Per-transaction access list
	accessList *accessList

	// Transient storage
	transientStorage transientStorage

	// Journal of state modifications. This is the backbone of
	// Snapshot and RevertToSnapshot.
	journal        *journal
	validRevisions []revision
	nextRevisionId int

	// Measurements gathered during execution for debugging purposes
	AccountReads         time.Duration
	AccountHashes        time.Duration
	AccountUpdates       time.Duration
	AccountCommits       time.Duration
	StorageReads         time.Duration
	StorageHashes        time.Duration
	StorageUpdates       time.Duration
	StorageCommits       time.Duration
	SnapshotAccountReads time.Duration
	SnapshotStorageReads time.Duration
	SnapshotCommits      time.Duration
	TrieDBCommits        time.Duration

	AccountUpdated int
	StorageUpdated int
	AccountDeleted int
	StorageDeleted int
}
```


### db

`StateDB` 结构中的第一个变量 `db` 是一个由 `Database` 类型定义的。这里的 `Database` 是一个抽象层的接口类型，它的定义如下所示。我们可以看到在`Database`接口中定义了一些操作更细粒度的数据管理模块的函数。例如 `DiskDB()` 函数会返回一个更底层的 key-value disk database 的实例，`TrieDB()` 函数会返回一个指向更底层的 Trie Databse 的实例。这两个模块都是非常重要的管理链上数据的模块。由于这两个模块本身就涉及到了大量的细节，因此我们在此就不对两个模块进行细节分析。在后续的章节中，我们会单独的对这两个模块的实现进行解读。

```go
type Database interface {
	// OpenTrie opens the main account trie.
	OpenTrie(root common.Hash) (Trie, error)

	// OpenStorageTrie opens the storage trie of an account.
	OpenStorageTrie(stateRoot common.Hash, addrHash, root common.Hash) (Trie, error)

	// CopyTrie returns an independent copy of the given trie.
	CopyTrie(Trie) Trie

	// ContractCode retrieves a particular contract's code.
	ContractCode(addrHash, codeHash common.Hash) ([]byte, error)

	// ContractCodeSize retrieves a particular contracts code's size.
	ContractCodeSize(addrHash, codeHash common.Hash) (int, error)

	// DiskDB returns the underlying key-value disk database.
	DiskDB() ethdb.KeyValueStore

	// TrieDB retrieves the low level trie database used for data storage.
	TrieDB() *trie.Database
}
```
### Trie

这里的 `trie` 变量同样的是由一个 `Trie` 类型的接口定义的。通过这个 `Trie` 类型的接口，上层其他模块就可以通过 `StateDB.tire` 来具体的对 `trie` 的数据进行操作。 

```go
type Trie interface {
	// GetKey returns the sha3 preimage of a hashed key that was previously used
	// to store a value.
	//
	// TODO(fjl): remove this when StateTrie is removed
	GetKey([]byte) []byte

	// TryGet returns the value for key stored in the trie. The value bytes must
	// not be modified by the caller. If a node was not found in the database, a
	// trie.MissingNodeError is returned.
	TryGet(key []byte) ([]byte, error)

	// TryGetAccount abstracts an account read from the trie. It retrieves the
	// account blob from the trie with provided account address and decodes it
	// with associated decoding algorithm. If the specified account is not in
	// the trie, nil will be returned. If the trie is corrupted(e.g. some nodes
	// are missing or the account blob is incorrect for decoding), an error will
	// be returned.
	TryGetAccount(address common.Address) (*types.StateAccount, error)

	// TryUpdate associates key with value in the trie. If value has length zero, any
	// existing value is deleted from the trie. The value bytes must not be modified
	// by the caller while they are stored in the trie. If a node was not found in the
	// database, a trie.MissingNodeError is returned.
	TryUpdate(key, value []byte) error

	// TryUpdateAccount abstracts an account write to the trie. It encodes the
	// provided account object with associated algorithm and then updates it
	// in the trie with provided address.
	TryUpdateAccount(address common.Address, account *types.StateAccount) error

	// TryDelete removes any existing value for key from the trie. If a node was not
	// found in the database, a trie.MissingNodeError is returned.
	TryDelete(key []byte) error

	// TryDeleteAccount abstracts an account deletion from the trie.
	TryDeleteAccount(address common.Address) error

	// Hash returns the root hash of the trie. It does not write to the database and
	// can be used even if the trie doesn't have one.
	Hash() common.Hash

	// Commit collects all dirty nodes in the trie and replace them with the
	// corresponding node hash. All collected nodes(including dirty leaves if
	// collectLeaf is true) will be encapsulated into a nodeset for return.
	// The returned nodeset can be nil if the trie is clean(nothing to commit).
	// Once the trie is committed, it's not usable anymore. A new trie must
	// be created with new root and updated trie database for following usage
	Commit(collectLeaf bool) (common.Hash, *trie.NodeSet)

	// NodeIterator returns an iterator that returns nodes of the trie. Iteration
	// starts at the key after the given start key.
	NodeIterator(startKey []byte) trie.NodeIterator

	// Prove constructs a Merkle proof for key. The result contains all encoded nodes
	// on the path to the value at key. The value itself is also included in the last
	// node and can be retrieved by verifying the proof.
	//
	// If the trie does not contain a value for key, the returned proof contains all
	// nodes of the longest existing prefix of the key (at least the root), ending
	// with the node that proves the absence of the key.
	Prove(key []byte, fromLevel uint, proofDb ethdb.KeyValueWriter) error
}
```

## StateDB 的持久化

当新的 Block 被添加到 Blockchain 时，State 的数据并不一会立即被写入到 Disk Database 中。在`writeBlockWithState`函数中，函数会判断 `gc` 条件，只有满足一定的条件，才会在此刻调用 `TrieDB` 中的 `Cap` 或者 `Commit` 函数将数据写入 Disk Database 中。

具体 World State 的更新顺序是:

```
StateDB --> Memory_Trie_Database --> LevelDB
```

StateDB 调用 `Commit` 的时候并没有同时触发 `TrieDB` 的 Commit 。

在Block被插入到 Blockchain 的这个Workflow中，stateDB的commit首先在`writeBlockWithState`函数中被调用了。之后`writeBlockWithState`函数会判断 `GC` 的状态来决定在本次调用中，是否需要向 `Disk Database` 写入数据。当新的Block被添加到Blockchain时，State 的数据并不一会立即被写入到 Disk Database 中。在`writeBlockWithState`函数中，函数会判断 `gc` 条件，只有满足一定的条件，才会在此刻调用 TrieDB 中的 `Cap` 或者 `Commit` 函数将数据写入Disk Database中。
