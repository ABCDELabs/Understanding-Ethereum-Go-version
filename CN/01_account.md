# Account and Contract

## Account数据结构分析

### Background

在本文中我们来探索一下以太坊中的基本数据元(Metadata)之一的Account。

我们知道，Ethereum是基于交易的状态机模型(Transaction-based State Machine)来运行的。在这种模型中，State基于Transaction的执行(数据更新/删除/创建)，而转移到另一个State。具体的说，Transaction的执行会让系统元对象(Meta Object)的数据值发生改变，表现为系统元对象从一个状态转换到另一个状态。在Ethereum中，这个元对象就是Account。State表现(represent)出来的是Account在某个时刻的包含/对应的数据的值。

- Account --> Object
- State   --> The value of the Object

In general, Account (账户)是参与链上交易的基本角色，是Ethereum状态机模型中的基本单位，承担了链上交易的发起者以及交易接收者的角色。

目前，在以太坊中，有两种类型的Account，分别是外部账户(EOA)以及合约(Contract)。

外部账户(EOA)由用户直接控制的账户，负责签名并发起交易(transaction)。

合约(Contract)由外部账户通过Transaction创建，用于在链上保存**不可篡改的**保存**图灵完备的代码段**，以及保存一些**持久化的数据**。这些代码段使用专用语言书写(Like: Solidity)，并且通常提供一些对外部访问API函数。这些函数通常用于计算以及查询或修改合约中的持久化数据。通常我们经常看到这样的描述"**一旦被记录到区块链上数据不可被修改**，或者**不可篡改的智能合约**"。现在我们知道这种描述是不准确。针对一个链上的智能合约，不可修改/篡改的部分是合约中的代码段，或说是合约中的*函数逻辑*/*代码逻辑是*不可以被修改/篡改的。而链上合约中的持久化的数据部分是可以通过调用代码段中的函数进行数据操作的(CURD)。用户在构造Transaction时只能调用一个合约中的API函数。如果一个用户只希望查询某些合约中的持久化数据，而不进行写操作的话，那么他不需要通过构造一个Transaction来查询数据。他可以通过直接调用本地数据中的对应的仅包含查询操作的函数代码或者请求其他节点存储的代码来操作。如果用户需要对合约中的数据进行更新，那么他就要构造一个Transaction来请求合约中相对应的函数。对于如何编写合约，以及Ethereum如何解析和执行Transaction调用的API的，Transaction的构造我们会在后面的文章中详细的进行解读。

### Account and stateObject

在实际代码中，这两种Account都是由stateObject这一结构定义的。stateObject的相关代码位于core/state/state_object.go文件中，隶属于package state。我们摘录了stateObject的结构代码，如下所示。通过下面的代码，我们可以观察到，stateObject是由小写字母开头。根据go语言的特性，我们可以知道这个结构主要用于package内部数据操作，并不对外暴露。

```go
  // stateObject represents an Ethereum account which is being modified.
  //
  // The usage pattern is as follows:
  // First you need to obtain a state object.
  // Account values can be accessed and modified through the object.
  // Finally, call CommitTrie to write the modified storage trie into a database.
  type stateObject struct {
    address  common.Address
    addrHash common.Hash // hash of ethereum address of the account
    data     types.StateAccount
    db       *StateDB
    dbErr error

    // Write caches.
    trie Trie // storage trie, which becomes non-nil on first access
    code Code // contract bytecode, which gets set when code is loaded

    // 这里的Storage 是一个 map[common.Hash]common.Hash
    originStorage  Storage // Storage cache of original entries to dedup rewrites, reset for every transaction
    pendingStorage Storage // Storage entries that need to be flushed to disk, at the end of an entire block
    dirtyStorage   Storage // Storage entries that have been modified in the current transaction execution
    fakeStorage    Storage // Fake storage which constructed by caller for debugging purpose.

    // Cache flags.
    // When an object is marked suicided it will be delete from the trie
    // during the "update" phase of the state transition.
    dirtyCode bool // true if the code was updated
    suicided  bool
    deleted   bool
  }
```

### Address

在stateObject这一结构体中，开头的两个成员变量为address以及address的哈希值addrHash。address是common.Address类型，address是common.Hash类型，它们分别对应了一个20字节长度的byte数组和一个32字节长度的byte数组。关于这两种数据类型的定义如下所示。

```go
// Lengths of hashes and addresses in bytes.
const (
 // HashLength is the expected length of the hash
 HashLength = 32
 // AddressLength is the expected length of the address
 AddressLength = 20
)
// Address represents the 20 byte address of an Ethereum account.
type Address [AddressLength]byte
// Hash represents the 32 byte Keccak256 hash of arbitrary data.
type Hash [HashLength]byte
```

在Ethereum中，每个Account都拥有独一无二的address，用于检索。Address作为每个Account的身份信息，类似于现实生活中的身份证，它与用户信息时刻绑定而且不能被修改。Ethereum通过Account Address来构建Merkle Patricia Trie来管理所有的Account state。MPT结构，也被称为World State Trie(or World State)。关于MPT结构以及World State的细节我们会在之后的文章中详细说明。

### data and StateAccount

继续向下探索，我们会遇到成员变量data，它是一个types.StateAccount类型的变量。在上面的分析中我们提到，stateObject这种类型只对Package State这个内部使用。所以相应的，Package State也为外部Package API提供了与Account相关的数据类型"State Account"。在上面的代码中我们就可以看到，"State Account"对应了State Object中"data Account"成员变量。State Account的具体数据结构的被定义在"core/types/state_account.go"文件中(~~在之前的版本中Account的代码位于core/account.go~~)，其定义如下所示。

```go
// Account is the Ethereum consensus representation of accounts.
// These objects are stored in the main account trie.
type StateAccount struct {
  Nonce    uint64
  Balance  *big.Int
  Root     common.Hash // merkle root of the storage trie
  CodeHash []byte
}
```

其中的包含四个变量为:

- Nonce 表示该账户发送的交易序号，随着账户发送的交易数量的增加而单调增加。
- Balance 表示该账户的余额。这里的余额指的是链上的Global Token Ether。
- Root 表示当前账户的下Storage层的 Merkle Patricia Trie的Root。EOA账户这个部分为空值。
- CodeHash是该账户的Contract代码的哈希值。EOA账户这个部分为空值。

### db

上述的几个成员变量基本覆盖了Account自身定义有关的全部成员变量。那么继续向下看，我们会遇到db和dbErr这两个成员变量。db这个变量保存了一个StateDB类型的指针(或者称为句柄handle)。这是为了方便调用StateDB相关的API对Account所对应的stateObject进行操作。StateDB本质上是Ethereum用于管理stateObject信息的而抽象出来的内存数据库，所有的Account数据的更新，检索都会使用StateDB提供的API。关于StateDB的具体实现，功能，以及如何与更底层(leveldb)进行结合的，我们会在之后的文章中进行详细描述。

### Cache

对于剩下的成员变量，它们的主要用于内存Cache。trie用于保存Contract中的持久化存储的数据，code用于缓存contract中的代码段到内存中，它是一个byte数组。剩下的四个Storage字段主要在执行Transaction的时候缓存Contract合约修改的持久化数据，比如dirtyStorage就用于缓存在Block被Finalize之前，Transaction所修改的合约中的持久化存储数据。对于外部账户，由于没有代码字段，所以对应stateObject对象中的code字段，以及四个Storage类型的字段对应的变量的值都为空(originStorage, pendingStorage, dirtyStorage, fakeStorage)。关于Contract的Storage层的详细信息，我们会在后面部分进行详细的描述。


## 深入Account

### Private Key & Public Kay & Address

#### 账户安全的问题

我们经常会在各种科技网站，自媒体上听到这样的说法，"用户在区块链系统中保存的Cryptocurrency/Token，除了用户自己，不存在一个中心化的第三方可以不经过用户的允许转走你的财富"。这个说法基本是正确的。目前，用户账户里的由链级别定义Crypto，或者称为原生货币(Native Token)，比如Ether，Bitcoin，BNB(Only in BSC)，是没办法被第三方在不被批准的情况下转走的。这是因为链级别上的所有数据的修改都要经过用户私钥(Private Key)签名的Transaction。只要用户保管好自己账户的私钥(Private Key)，保证其没有被第三方知晓，就没有人可以转走你链上的财富。

我们说上述说法是基本正确，而不是完全正确的原因有两个。首先，用户的链上数据安全是基于当前Ethereum使用的密码学工具足够保证：不存在第三方可以在**有限的时间**内在**不知道用户私钥的前提**下获取到用户的私钥信息来伪造签名交易。当然这个安全保证前提是当今Ethereum使用的密码学工具的强度足够大，没有计算机可以在有限的时间内hack出用户的私钥信息。在量子计算机出现之前，目前Ethereum和其他Blockchain使用的密码学工具的强度都是足够安全的。这也是为什么很多新的区块链项目在研究抗量子计算机密码体系的原因。第二点原因是，当今很多的所谓的Crypto/Token并不是链级别的数据，而是在链上合约中存储的数据，比如ERC-20 Token和NFT对应的ERC-721的Token。由于这部分的Token都是基于合约代码生成和维护的，所以这部分Token的安全依赖于合约本身的安全。如果合约本身的代码是有问题的，存在后门或者漏洞，比如存在给第三方任意提取其他账户下Token的漏洞，那么即使用户的私钥信息没有泄漏，合约中的Token仍然可以被第三方获取到。由于合约的代码段在链上是不可修改的，合约代码的安全性是极其重要的。所以，有很多研究人员，技术团队在进行合约审计方面的工作，来保证上传的合约代码是安全的。此外随着Layer-2技术和一些跨链技术的发展，用户持有的“Token”，在很多情况下不是我们上面提到的安全的Naive Token，而是ERC-20甚至只是其他合约中的简单数值记录。这种类型的资产的安全性是低于layer-1上的Native Token的。用户在持有这类资产的时候需要小心。这里我们推荐阅读Jay Freeman所分析的关于一个热门Layer-2系统Optimism上的由于非Naive Token造成的[任意提取漏洞](https://www.saurik.com/optimism.html)。

#### Account Generation

下面我们简单讲述，在Ethereum中一个账户的私钥和地址是如何产生的。

- 首先我们通过随机得到一个长度64位account的私钥。这个私钥就是平时需要用户激活钱包时需要的记录，一旦这个私钥暴露了，钱包也将不再安全。
  - 64个16进制位，256bit，32字节
    `var AlicePrivateKey = "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"`

- 在得到私钥后，我们使用私钥来计算公钥和account的地址。基于上述私钥，我们使用ECDSA算法，选择spec256k1曲线进行计算。通过将私钥带入到所选择的椭圆曲线中，计算出点的坐标即是公钥。以太坊和比特币使用了同样的spec256k1曲线，在实际的代码中，我们也可以看到在crypto中，go-Ethereum直接调用了比特币的代码。
    `ecdsaSK, err := crypto.ToECDSA(privateKey)`

- 对私钥进行椭圆加密之后，我们可以得到64bytes的数，它是由两个32bytes的数构成，这两个数代表了spec256k1曲线上某个点的XY值。
    `ecdsaPK := ecdsaSK.PublicKey`
- 以太坊的地址，是基于上述公钥(ecdsaSK.PublicKey)的 [Keccak-256算法] 之后的后20个字节，并且用0x开头。
  - Keccak-256是SHA-3（Secure Hash Algorithm 3）标准下的一种哈希算法
    `addr := crypto.PubkeyToAddress(ecdsaSK.PublicKey)`

#### Signature & Verification

- Hash（m,R）*X +R = S* P
- P是椭圆曲线函数的基点(base point) 可以理解为一个P是一个在曲线C上的一个order 为n的加法循环群的生成元. n为质数。
- R = r * P (r 是个随机数，并不告知verifier)
- 以太坊签名校验的核心思想是:首先基于上面得到的ECDSA下的私钥ecdsaSK对数据msg进行签名(sign)得到msgSig.
    `sig, err := crypto.Sign(msg[:], ecdsaSK)`
    `msgSig := decodeHex(hex.EncodeToString(sig))`

- 然后基于msg和msgSig可以反推出来签名的公钥（用于生成账户地址的公钥ecdsaPK）。
    `recoveredPub, err := crypto.Ecrecover(msg[:],msgSig)`
- 通过反推出来的公钥得到发送者的地址，并与当前txn的发送者在ECDSA下的pk进行对比。
    `crypto.VerifySignature(testPk, msg[:], msgSig[:len(msgSig)-1])`
- 这套体系的安全性保证在于，即使知道了公钥ecdsaPk/ecdsaSK.PublicKey也难以推测出 ecdsaSK以及生成他的privateKey。

#### ECDSA & spec256k1曲线

- Elliptic curve point multiplication
  - Point addition P + Q = R
  - Point doubling P + P = 2P
- y^2 = x^3 +7
- Based Point P是在椭圆曲线上的群的生成元
- x次computation on Based Point得到X点，x为私钥，X为公钥。x由Account Private Key得出。
- 在ECC中的+号不是四则运算中的加法，而是定义椭圆曲线C上的新的二元运算(Point Multiplication)。他代表了过两点P和Q的直线与椭圆曲线C的交点R‘关于X轴对称的点R。因为C是关于X轴对称的所以关于X对称的点也都在椭圆曲线上。

## 深入Contract

- 这部分的示例代码位于: [[example/signature](example/signature)]中。

### Contract Storage (合约存储)

[在文章的开头](#general Background)我们提到，在外部账户对应的，stateObject结构体的实例中，有四个Storage类型的变量是空值。那显然的，这四个变量是为Contract类型的账户准备的。

在"state_object.go"文件的开头部分(41行左右)，我们可以找到Storage类型的定义。具体如下所示。

```go
type Storage map[common.Hash]common.Hash
```

我们可以看到，*Storage*是一个key和value都是common.Hash类型的map结构。common.Hash类型，则对应了一个长度为32bytes的byte类型数组。这个类型在go-ethereum中被大量使用，通常用于表示32字节长度的数据，比如Keccak256的哈希值。在之后的旅程中，我们也会经常看到它的身影，它的定义在common.type.go文件中。

```go
// HashLength is the expected length of the hash
HashLength = 32
// Hash represents the 32 byte Keccak256 hash of arbitrary data.
type Hash [HashLength]byte
```

EOA与Contract不同的点在于，EOA并没有维护自己的Storage层以及代码(codeHash)。相比与外部账户，Contract账户额外保存了一个存储层(Storage)用于存储合约代码中持久化的变量的数据。而上面的我们提到的stateObject中的四个Storage类型的变量，就是作为Contract Storage层的内存缓存。

Storage层的基本组成单元称为槽 (Slot)。每个Slot的大小是256bits，最多保存32 bytes的数据。作为基本的存储单元，Slot类似于内存的page以及HDD中的Block，可以通过地址索引的方式被上层函数访问(state_object/getState())。Slot的索引key的长度同样是32 bytes(256 bits)，寻址空间从0x0000000000000000000000000000000000000000000000000000000000000000 到 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF。因此，每个Contract的Storage层最多可以保存$2^{256} - 1$个Slot。合约帐户同样使用MPT，作为可验证的索引结构来管理Slot。Storage Trie的根数据被保存在StateAccount结构体中的Root变量中，它是一个32bytes长的byte数组。

### Contract Storage Example One

我们使用一个简单的合约来展示Contract Storage层的逻辑，合约代码如下所示。在本例中，Storage合约保存了三个持久化uint256 变量(number, number1, and number2)，并通过stores函数给它们进行赋值。

```solidity
// SPDX-License-Identifier: GPL-3.0

pragma solidity >=0.7.0 <0.9.0;

/**
 * @title Storage
 * @dev Store & retrieve value in a variable
 */
contract Storage {

    uint256 number;
    uint256 number1;
    uint256 number2;

    function stores(uint256 num) public {
        number = num;
        number1 = num + 1;
        number2 = num + 2;
    }
    
    function get_number() public view returns (uint256){
        return number;
    }
    
    function get_number1() public view returns (uint256){
        return number1;
    }
    
    function get_number2() public view returns (uint256){
        return number2;
    }
}
```

我们使用remix来在本地部署这个合约，并使用remix debugger构造transaction调用stores(1)函数，并观察Storage层的变化。在Transaction生效之后，合约中三个变量的值将被分别赋给1，2，3。此时，我们观察Storage层会发现，现在的存储层增加了三个Storage Object，或者说使用了三个Slots。每个Object包含一个256 bits的key和256 bits的value字段（本例中表现为64位的16进制数）。其中Key的值是从0开始的递增整数，它代表了Slot的索引值(或者该Slot在Storage层对应的物理位置)。它们的value存储了合约中三个变量值(1,2,3)。此外，每个object外层的值，则是key值的sha3的哈希值，比如下面的"0x290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563" 是keccak(0)的值，"0xb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6" 对应了是keccak(1)的值 。我们在示例代码中展示了这一结果。这个hash值会被作为参数，用在state_object/getState()函数中。

```json
{
 "0x290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563": {
  "key": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "value": "0x0000000000000000000000000000000000000000000000000000000000000001"
 },
 "0xb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6": {
  "key": "0x0000000000000000000000000000000000000000000000000000000000000001",
  "value": "0x0000000000000000000000000000000000000000000000000000000000000002"
 },
 "0x405787fa12a823e0f2b7631cc41b3ba8828b3321ca811111fa75cd3aa3bb5ace": {
  "key": "0x0000000000000000000000000000000000000000000000000000000000000002",
  "value": "0x0000000000000000000000000000000000000000000000000000000000000003"
 }
}
```

### Account Storage Example Two

值得注意的是，如果我们调整一下合约中变量的声明顺序，从(number，number1，number2)调整为(number 2, number 1, number)，则会在Storage 层观察到不一样的结果。

```solidity
// SPDX-License-Identifier: GPL-3.0

pragma solidity >=0.7.0 <0.9.0;

/**
 * @title Storage
 * @dev Store & retrieve value in a variable
 */
contract Storage {

    uint256 number2;
    uint256 number;
    uint256 number;

    function stores(uint256 num) public {
        number = num;
        number1 = num + 1;
        number2 = num + 2;
    }
    
    function get_number() public view returns (uint256){
        return number;
    }
    
    function get_number1() public view returns (uint256){
        return number1;
    }
    
    function get_number2() public view returns (uint256){
        return number2;
    }
}
```

我们可以发现number2的结果被存储在了第一个Slot中（Key:"0x0000000000000000000000000000000000000000000000000000000000000000"），而number的值北存储在了第三个Slot中 (Key:"0x0000000000000000000000000000000000000000000000000000000000000002")。

```json
{
  "0x290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563": {
    "key": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "value": "0x0000000000000000000000000000000000000000000000000000000000000003"
    },
  "0xb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6": {
    "key": "0x0000000000000000000000000000000000000000000000000000000000000001",
    "value": "0x0000000000000000000000000000000000000000000000000000000000000002"
  },
  "0x405787fa12a823e0f2b7631cc41b3ba8828b3321ca811111fa75cd3aa3bb5ace": {
    "key": "0x0000000000000000000000000000000000000000000000000000000000000002",
    "value": "0x0000000000000000000000000000000000000000000000000000000000000001"
  }
}
```

这个实验可以证明，在Ethereum中，变量对应的存储层的Slot，是按照其在在合约中的声明顺序，从第一个Slot（Key：0）开始分配的。

### Account Storage Example Three

我们再考虑另一种情况：声明的三个变量，但只对其中的两个变量进行赋值。具体的来说，我们按照number，number1，和number2的顺序声明三个uint256变量。但是，在函数stores中只对number1和number2进行赋值操作。合约代码如下所示。

```solidity
// SPDX-License-Identifier: GPL-3.0

pragma solidity >=0.7.0 <0.9.0;

/**
 * @title Storage
 * @dev Store & retrieve value in a variable
 */
contract Storage {
    uint256 number;
    uint256 number1;
    uint256 number2;

    function stores(uint256 num) public {
        number1 = num + 1;
        number2 = num + 2;
    }
    
    function get_number() public view returns (uint256){
        return number;
    }
    
    function get_number1() public view returns (uint256){
        return number1;
    }
    
    function get_number2() public view returns (uint256){
        return number2;
    }
}
```

基于上述合约，我们构造transaction 并调用stores函数，输入参数1，将number1和number2的值修改为2，和3。在transaction执行完成后，我们可以观察到Storage层Slot的结果如下所示。

```json
{
 "0xb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6": {
  "key": "0x0000000000000000000000000000000000000000000000000000000000000001",
  "value": "0x0000000000000000000000000000000000000000000000000000000000000002"
 },
 "0x405787fa12a823e0f2b7631cc41b3ba8828b3321ca811111fa75cd3aa3bb5ace": {
  "key": "0x0000000000000000000000000000000000000000000000000000000000000002",
  "value": "0x0000000000000000000000000000000000000000000000000000000000000003"
 }
}
```

我们可以看到，transaction的执行只对在合约的Storage中位置在1和2位置的两个Slot进行了赋值。值得注意的是，在本例中，针对Slot的赋值是从1号位置Slot的开始，而不是0号Slot。这说明，对于固定长度的变量，其值的所占用的Slot的位置在Contract初始化开始的时候就已经分配的。即使变量只是被声明没有真正的赋值，其对应的保存值的Slot已经被分配好了。而不是在第一次给变量赋值的时候，进行再对变量的Slot值进行分配。

![Remix Debugger](../figs/01/remix.png)

### Account Storage Example Four

在Solidity中，有一类特殊的类型**Address**，用于表示账户的地址信息。例如在ERC-20合约中，所有用户拥有的token信息是被存储在一个(address->uint)的map结构中。这个map的key是Address类型的，它表示了用户实际的address。目前Address的大小为160bits(20bytes)，并不足以填满一整个Slot。因此当Address作为value单独存储在的时候，它并不会排他的独占用一个Slot。我们使用下面的例子来说明。

在下面的示例中，我们声明了三个变量，分别是number(uint256)，addr(address)，以及isTrue(bool)。我们知道，在以太坊中Address是一个长度为20 bytes的字符串，所以一个Address类型是没办法填满整个的Slot的。布尔类型在以太坊中只需要一个bit(0 or 1)就可以表示. 我们构造transaction调用函数storeaddr。函数的input为1 “0xb6186d3a3D32232BB21E87A33a4E176853a49d12”。

```solidity
// SPDX-License-Identifier: GPL-3.0

pragma solidity >=0.7.0 <0.9.0;

/**
 * @title Storage
 * @dev Store & retrieve value in a variable
 */
contract Storage {


    uint256 number;
    address addr;
    bool isTrue;


    function stores(uint256 num) public {
        // number1 = num + 1;
        // number2 = num + 2;
    }
    
    function storeaddr(uint256 num, address a) public {
        number = num;
        addr = a;
        isTure = true;
    }
    
    function get_number() public view returns (uint256){
        return number;
    }
    
}
```

Transaction的运行后的结果如下面的Json所示。我们可以观察到，在本例中Contract声明了三个变量但是Storage只占用了两个Slot。按照我们上面的发现，在第二个slot(Key:0x0000000000000000000000000000000000000000000000000000000000000001)保存了addr和isTrue的值。

```json
{
 "0x290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563": {
  "key": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "value": "0x0000000000000000000000000000000000000000000000000000000000000001"
 },
 "0xb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6": {
  "key": "0x0000000000000000000000000000000000000000000000000000000000000001",
  "value": "0x000000000000000000000001b6186d3a3d32232bb21e87a33a4e176853a49d12"
 }
}
```

### Account Storage Example Five

对于变长数组，map结构的存储构造则更为复杂。虽然Map本身就是key-value的结构，但是在Storage 层并不直接使用map中key的值或者key的值的hash值来作为Storage的索引值。目前，使用map的key的值和当前数组所在变量声明位置对应的slot的值进行拼接，再进行keccak256哈希值作为索引。我们在下面的例子中展示了EVM是如何处理mapping这种变长的数据结构的。在下面的合约中，我们声明了几个定长的uint256类型的对象，和一个string=>uint256类型的Mapping对象。

<!-- Todo: 变长数据结构的存储情况。 -->

```solidity
// SPDX-License-Identifier: GPL-3.0

pragma solidity >=0.7.0 <0.9.0;

/**
 * @title Storage
 * @dev Store & retrieve value in a variable
 */
contract Storage {
    
    uint256 number;
    
    mapping(string => uint256) balances;

    function set_balance(uint256 num) public {
        number = num;
        balances["hsy"] = num;
        balances["lei"] = num + 1;
    }
    
    function get_number() public view returns (uint256){
        return number;
    }
    
}
```

我们发现，对于定长变量number被存储在了第一个Slot(key:0x0000000000000000000000000000000000000000000000000000000000000000)中。但是对于mapping变量balances，它包含的两个数据并没有按照变量定义的物理顺序来定义Slot。此外，我们也观察到存储这两个值的Slot的key，也并不是这两个字在mapping中key的直接hash。这是由于Solidity对这种变长的数据结构有额外的分配Slot的方式。具体的来说Solidity会使用mapping中元素的key值与，当前mapping本身对应的slot的位置进行拼接，之后再进行其使用keccak256的hash来得到map中元素最终的存储位置。比如在本例中，按照变量定义的顺序，balances这个mapping会被分配到第二个Slot，对应的绝对位置1。那么balances中的kv对分配到的slot的位置就是，keccak(key, 1)，这里是一个特殊的拼接操作。

```json
{
 "0x290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563": {
  "key": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "value": "0x0000000000000000000000000000000000000000000000000000000000000001"
 },
 "0xa601d8e9cd2719ca27765dc16042655548d1ac3600a53ffc06b4a06a12b7c65c": {
  "key": "0xbaded3bf529b04b554de2e4ee0f5702613335896b4041c50a5555b2d5e279f91",
  "value": "0x0000000000000000000000000000000000000000000000000000000000000001"
 },
 "0x53ac6681d92653b13055d2e265b672e2db2b2a19407afb633928597f144edbb0": {
  "key": "0x56a8a0d158d59e2fd9317c46c65b1e902ed92f726ecfe82c06c33c015e8e6682",
  "value": "0x0000000000000000000000000000000000000000000000000000000000000002"
 }
}
```

我们通过go语言来调用相关的库来验证一下上面的结论。对于 balances["hsy"]，它应该分配的slot的位置可以由下面的代码求得:

```go
k1 := solsha3.SoliditySHA3([]byte("hsy"), solsha3.Uint256(big.NewInt(int64(1))))
fmt.Printf("Test the Solidity Map storage Key1:         0x%x\n", k1)
```

这里的k1是一个意义是一个数值，代表了slot的在storage层的绝对位置。

## Wallet

- KeyStore
- Private Key
- 助记词

## Reference

- <https://www.freecodecamp.org/news/how-to-generate-your-very-own-bitcoin-private-key-7ad0f4936e6c/>
