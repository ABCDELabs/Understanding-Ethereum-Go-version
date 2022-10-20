# Account and Contract

## 概述

我们常常听到这样一个说法，"Ethereum 和 Bitcoin 最大的不同之一是二者使用链上数据模型不同。其中，Bitcoin 是基于 UTXO 模型的 Blockchain/Ledger 系统，Ethereum是基于 Account/State 模型的系统"。那么，这个另辟蹊径的 Account/State 模型究竟不同在何处呢？在本文，我们就来探索一下以太坊中的基本数据单元(Metadata)之一的`Account`。

简单的来说，Ethereum 的运行是一种*基于交易的状态机模型*(Transaction-based State Machine)。整个系统由若干的账户组成 (Account)，类似于银行账户。状态(State)反应了某一账户(Account)在*某一时刻*下的值(value)。在以太坊中，State 对应的基本数据结构，称为 StateObject。当 StateObject 的值发生了变化时，我们称为*状态转移*。在 Ethereum 的运行模型中，StateObject 所包含的数据会因为 Transaction 的执行引发数据更新/删除/创建，引发状态转移，我们说：StateObject 的状态从当前的 State 转移到另一个 State。

在 Ethereum 中，承载 StateObject 的具体实例就是 Ethereum 中的 Account。通常，我们提到的 State 具体指的就是 Account 在某个时刻的包含的数据的值。

- Account --> StateObject
- State   --> The value/data of the Account

总的来说, Account (账户)是参与链上交易(Transaction)的基本角色，是Ethereum状态机模型中的基本单位，承担了链上交易的发起者以及接收者的角色。目前，在以太坊中，有两种类型的Account，分别是外部账户(EOA)以及合约账户(Contract)。

### EOA

外部账户(EOA)是由用户直接控制的账户，负责签名并发起交易(Transaction)。用户通过Account 的私钥来保证对账户数据的控制权。

合约账户(Contract)，简称为合约，是由外部账户通过Transaction创建的。合约账户，保存了**不可篡改的图灵完备的代码段**，以及一些**持久化的数据变量**。这些代码使用专用的图灵完备的编程语言编写(Solidity)，并通常提供一些对外部访问 API 接口函数。这些 API 接口函数可以通过构造 Transaction，或者通过本地/第三方提供的节点 RPC 服务来调用。这种模式构成了目前的 DApp 生态的基础。

通常，合约中的函数用于计算以及查询或修改合约中的持久化数据。我们经常看到这样的描述"**一旦被记录到区块链上数据不可被修改**，或者**不可篡改的智能合约**"。现在我们知道这种笼统的描述其实是不准确。针对一个链上的智能合约，不可修改/篡改的部分是合约中的代码段，或说合约中的*函数逻辑*/*代码逻辑是*不可以被修改/篡改的。而合约中的**持久化的数据变量**是可以通过调用代码段中的函数进行数据操作的(CURD)。具体的操作方式取决于合约函数中的代码逻辑。

根据*合约中函数是否会修改合约中持久化的变量*，合约中的函数可以分为两种: *只读函数*和*写函数*。如果用户**只**希望查询某些合约中的持久化数据，而不对数据进行修改的话，那么用户只需要调用相关的只读函数。调用只读函数不需要通过构造一个 Transaction 来查询数据。用户可以通过直接调用本地节点或者第三方节点提供的 RPC 接口来直接调用对应的合约中的*只读函数*。如果用户需要对合约中的数据进行更新，那么他就要构造一个Transaction 来调用合约中相对应的*写函数*。注意，每个 Transaction 每次调用一个合约中的一个*写函数*。因为，如果想在链上实现复杂的逻辑，需要将*写函数*接口化，在其中调用更多的逻辑。

对于如何编写合约，以及 Ethereum 的执行层如何解析 Transaction 并调用对应的合约中函数的，我们会在后面的文章中详细的进行解析。

## StateObject, Account, Contract

### 概述

在实际代码中，这两种 Account 都是由`stateObject`这一数据结构定义的。`stateObject`的相关代码位于*core/state/state_object.go*文件中，隶属于*package state*。我们摘录了`stateObject`的结构代码，如下所示。通过下面的代码，我们可以观察到，`stateObject`是由小写字母开头。根据 go 语言的特性，我们可以知道这个结构主要用于 package 内部数据操作，并不对外暴露。

```go
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

在`stateObject`这一结构体中，开头的两个成员变量为`address`以及 address 的哈希值`addrHash`。`address`是common.Address类型，`addrHash`是common.Hash类型，它们分别对应了一个**20字节**长的byte类型数组和一个32字节长的byte类型数组。关于这两种数据类型的定义如下所示。

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

在Ethereum中，每个 Account 都拥有独一无二的地址。Address作为每个 Account 的身份信息，类似于现实生活中的身份证，它与用户信息时刻绑定而且不能被修改。

### data and StateAccount

继续向下探索，我们会遇到成员变量 `data`，它是一个`types.StateAccount`类型的变量。在上面的分析中我们提到，`stateObject`这种类型只对 Package State 这个内部使用。所以相应的，Package State也为外部Package API提供了与Account相关的数据类型"State Account"。在上面的代码中我们就可以看到，"State Account"对应了State Object中" data Account" 成员变量。State Account的具体数据结构的被定义在"core/types/state_account.go"文件中(~~在之前的版本中Account的代码位于core/account.go~~)，其定义如下所示。

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

- Nonce 表示该账户发送的交易序号，随着账户发送的交易数量的增加而单调增加。每次发送一个交易，Nonce 的值就会加1。
- Balance 表示该账户的余额。这里的余额指的是链上的 Native Token Ether (以太)。
- Root 表示当前账户的下 Storage 层的 Merkle Patricia Trie的 Root。这里的存储层是为了管理合约中持久化变量准备的。对于 EOA账户这个部分为空值。
- CodeHash是该账户的Contract代码的哈希值。同样的，这个变量是用于保存合约账户中的代码的 hash ，EOA账户这个部分为空值。

### db

上述的几个成员变量基本覆盖了 Account 主生命周期相关的全部成员变量。那么我们继续向下看，会遇到`db`和`dbErr`这两个成员变量。db这个变量保存了一个 `StateDB` 类型的指针。这是为了方便调用 `StateDB` 相关的API对Account所对应的 `stateObject` 进行操作。StateDB本质上是用于管理`stateObject`信息的而抽象出来的内存数据库。所有的Account 数据的更新，检索都会使用 StateDB 提供的API。关于 StateDB 的具体实现，功能，以及如何与更底层物理存储层(leveldb)进行结合的，我们会在之后的文章中进行详细描述。

### Cache

对于剩下的成员变量，它们的主要用于内存Cache。trie用于保存和管理合约账户中的持久化变量存储的数据，code用于缓存合约中的代码段到内存中，它是一个byte数组。剩下的四个Storage 字段主要在执行 Transaction 的时候缓存合约修改的持久化数据，比如dirtyStorage 就用于缓存在 Block 被 Finalize 之前，Transaction所修改的合约中的持久化存储数据。对于外部账户，由于没有代码字段，所以对应 stateObject 对象中的code 字段，以及四个 Storage 类型的字段对应的变量的值都为空(originStorage, pendingStorage, dirtyStorage, fakeStorage)。

从调用关系上看，这四个缓存变量的修改顺序是: originStorage --> dirtyStorage--> pendingStorage。关于合于中的 Storage 层的详细信息，我们会在后面部分进行详细的描述。

## 深入Account

### 谁掌握了你的账户

我们经常会在各种科技网站/自媒体上看到这样的说法，"用户在区块链系统中保存的Cryptocurrency/Token，除了用户自己，不存在第三方可以不经过用户的允许转走你的财富"。这个说法基本是正确的。目前，用户账户里的由链级别定义的 Crypto/Token，或者称为原生货币(Native Token)，比如Ether，Bitcoin，BNB(Only in BSC)，是没办法被第三方在不被批准的情况下转走的。这是因为链级别上的所有数据的修改都要执行由用户私钥(Private Key)签名的Transaction。因此，只要用户保管好自己账户的私钥(Private Key)，保证其没有被第三方知晓，就没有人可以转走你链上的财富。

我们说上述说法是基本正确，而不是完全正确。原因有两个。首先，用户的链上数据安全是基于当前Ethereum使用的密码学工具足够保证：不存在第三方可以在**有限的时间**内在**不知道用户私钥的前提**下获取到用户的私钥信息来伪造签名交易。当然这个安全保证前提是当今Ethereum使用的密码学工具的强度足够大，没有计算机可以在有限的时间内hack出用户的私钥信息。在量子计算机出现之前，目前Ethereum和其他Blockchain使用的密码学工具的强度都是足够安全的。这也是为什么很多新的区块链项目在研究抗量子计算机密码体系的原因。第二点原因是，当今很多的所谓的Crypto/Token并不是链级别的代币，而是保存在合约中持久化变量中的数据，比如 ERC-20 Token 和NFT对应的 ERC-721 的Token。由于这部分的Token都是基于合约代码生成和维护的，所以这部分Token的安全依赖于合约本身的安全。如果合约本身的代码是有问题的，存在后门或者漏洞，比如存在给第三方任意提取其他账户下Token的漏洞。那么即使用户的私钥信息没有泄漏，合约中的Token仍然可以被第三方获取到。由于合约的代码段在链上是不可修改的，因此合约代码的安全性是极其重要的。目前有很多研究人员，技术团队在进行合约审计方面的工作，来保证上传的合约代码是安全的。随着Layer-2技术和一些跨链技术的发展，用户持有的`Token`，在很多情况下不是我们上面提到的安全的 Naive Token，而是 ERC-20 甚至只是其他合约中的简单数值记录。这种类型的资产的安全性是远低于低于layer-1上的 Native Toke n的。用户在持有这类资产的时候需要小心。这里我们推荐阅读 Jay Freeman 所分析的关于一个热门Layer-2系统Optimism上的由于非Naive Token造成的[任意提取漏洞](https://www.saurik.com/optimism.html)。

### Account Generation

首先，EOA账户的创建分为本地创建和链上注册两个部分。当我们使用诸如 Metamask 等钱包工具创建账户的时候，在区块链上并没有同步注册账户信息。链上账户的创建和管理都是通过`StateDB`模块来操作的，因此我们将`geth`中账户管理部分的代码整合到`StateDB`模块章节来一起讲述。而合约账户，或者说智能合约的创建是需要通过 EOA 账户构造特定的交易生成的。关于这部分的细节我们也放在之后的章节中进行解析。

下面我们简单分析一下，如何在本地创建一个 EOA 账户的。

总的来说，创建新账户的依赖的入口函数`NewAccount`位于 `accounts/keystore/keystore.go`文件中。函数有一个string类型的passphrase参数。注意，这个参数仅用于加密本地保存私钥的Keystore文件，与生成账户的私钥，地址的生成都无关。

```go
// passphrase 参数用于本地加密
func (ks *KeyStore) NewAccount(passphrase string) (accounts.Account, error) {
//生成account的函数
 _, account, err := storeNewKey(ks.storage, crand.Reader, passphrase)
 if err != nil {
  return accounts.Account{}, err
 }
 // Add the account to the cache immediately rather
 // than waiting for file system notifications to pick it up.
 ks.cache.add(account)
 ks.refreshWallets()
 return account, nil
}
```

上述代码段中，最核心的调用是`storeNewKey`函数。在`storeNewKey`函数中，首先就调用了`newKey`函数，该函数的主要功能就是生成一个账户需要的私钥和公钥对。而`newKey`函数的核心是调用了生成椭圆曲线加密对相关的函数`ecdsa.GenerateKey`。

```go
func newKey(rand io.Reader) (*Key, error) {
 privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand)
 if err != nil {
  return nil, err
 }
 return newKeyFromECDSA(privateKeyECDSA), nil
}

```

这部分的代码位于`crypto/ecdsa.go`文件中。由于这一部分涉及到了大量的椭圆曲线加密的知识，与以太坊的主要业务关系不大，因此对于这部分的内容，我们仅简述主要流程，对于数学具体原理在此我们不进行详细说明，感兴趣的读者可以自行搜索。值得注意的时候，在整个流程中，首先生成的是账户的私钥，而账户对应的地址，是基于该私钥在椭圆曲线上对用的公钥值，经过哈希计算得到的。

- 首先，我们在创建一个新的 EOA 账户的时候，我们首先会通过`GenerateKey`函数随机的得到一串私钥，它是一个 32bytes长的变量，通常表现为64位16进制数。这个私钥就是平时需要用户激活钱包时，发送交易时的门禁卡，一旦这个私钥暴露了，钱包也将不再安全。
  - 64个16进制位，256bit，32字节
    `var AlicePrivateKey = "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"`

- 在得到私钥后，我们使用私钥来计算公钥和地址地址。基于上述私钥，我们使用ECDSA算法，选择spec256k1曲线进行计算。通过将私钥带入到所选择的椭圆曲线中，计算出点的坐标即是公钥。以太坊和比特币使用了同样的spec256k1曲线，在实际的代码中，我们也可以看到在 go-Ethereum 直接调用了比特币的 secp256k1 的C语言代码。
    `ecdsaSK, err := crypto.ToECDSA(privateKey)`
- 对私钥进行椭圆加密之后，我们可以得到一个 64bytes 的数，它是由两个 32bytes 的数构成，这两个数代表了spec256k1曲线上某个点的XY值。
    `ecdsaPK := ecdsaSK.PublicKey`
- 以太坊的地址，是基于上述公钥(ecdsaSK.PublicKey)进行 **Keccak-256算法** 计算之后哈希值的后20个字节，用0x开头表示。
  - Keccak-256是SHA-3（Secure Hash Algorithm 3）标准下的一种哈希算法
    `addr := crypto.PubkeyToAddress(ecdsaSK.PublicKey)`

#### Signature & Verification

<!-- TODO: 这里写的不好，需要再改改 -->

- Hash（m,R）*X +R = S* P
- P是椭圆曲线函数的基点(base point) 可以理解为一个P是一个在曲线C上的一个order 为n的加法循环群的生成元，n是一个大质数。
- R = r * P (r 是个随机数，并不告知verifier)
- 以太坊签名校验的核心思想是:首先基于上面得到的ECDSA下的私钥ecdsaSK对数据 msg 进行签名(sign)得到msgSig.
    `sig, err := crypto.Sign(msg[:], ecdsaSK)`
    `msgSig := decodeHex(hex.EncodeToString(sig))`
- 然后基于 msg 和 msgSig 可以反推出来签名的公钥（用于生成账户地址的公钥ecdsaPK）。
    `recoveredPub, err := crypto.Ecrecover(msg[:],msgSig)`
- 通过反推出来的公钥可以得到发送者的地址，并与当前交易的发送者在 ECDSA 下的pk进行对比。
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

<!-- TODO: 这部分未来会整合到EVM章节中 -->

[在文章的开头](#general Background)我们提到，在外部账户对应的stateObject结构体的实例中，有四个Storage类型的变量是空值。那显然的，这四个变量是为Contract类型的账户准备的。

在*state_object.go*文件的开头部分(41行左右)，我们可以找到Storage类型的定义。具体如下所示。

```go
type Storage map[common.Hash]common.Hash
```

我们可以看到，*Storage*是一个 key 和 value 都是`common.Hash`类型的 map 结构，common.Hash 类型，则对应了一个长度为 32bytes 的 byte类型数组。这个类型在go-ethereum中被大量使用，通常用于表示32字节长度的数据，比如 Keccak256 函数的哈希值。在之后的旅程中，我们也会经常看到它的身影，它的定义在`common.type.go`文件中。

```go
// HashLength is the expected length of the hash
HashLength = 32
// Hash represents the 32 byte Keccak256 hash of arbitrary data.
type Hash [HashLength]byte
```

从功能层面讲，外部账户(EOA)与合约账户(Contract)不同的点在于: 外部账户并没有维护自己的代码(codeHash)以及额外的Storage层。相比与外部账户，合约账户额外保存了一个存储层(Storage)用于存储合约代码中持久化的变量的数据。在上文中我们提到，StateObject中的声明的四个 Storage 类型的变量，就是作为 Contract Storage 层的内存缓存。

在Ethereum中，每个合约都维护了自己的*独立*的 Storage 空间，我们称为 Storage 层。Storage 层的基本组成单元称为槽 (Slot)，若干个Slot按照*Stack*的方式集合在一起构造成了 Storage 层。每个Slot的大小是 256 bits，也就是最多保存32 bytes的数据。作为基本的存储单元，Slot管理的方式与内存或者HDD中的基本单元的管理方式类似，通过地址索引的方式被上层函数访问。Slot的地址索引的长度同样是32 bytes(256 bits)，寻址空间从 0x0000000000000000000000000000000000000000000000000000000000000000 到 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF。因此，每个Contract的Storage层最多可以保存$2^{256} - 1$个 Slot。也就说在理论状态下，一个Contract可以最多保存$(2^{256} - 1)$ bytes的数据，这是个相当大的数字。Contract同样使用 MPT 来管理 Storage 层的Slot。值得注意的是，Storage层的数据并不会被打包进入 Block 中。正如我们之前提到的，管理合约账户中 Storage 层 Storage Trie 的根数据被保存在 StateAccount 结构体中的 Root 变量中(它是一个32bytes长的byte数组)。当某个Contract的Storage层的数据发生变化时，根据骨牌效应，向上传导到整个的World State Root的值发生变化，从而影响到Chain链上数据。目前，Storage 层的数据读取和修改是在执行相关Transaction的时候，通过EVM调用两个专用的指令*OpSload*和*OpSstore*触发。

我们知道目前Ethereum中的大部分合约都通过 Solidity 语言编写。Solidity 做为强类型的图灵完备的语言，支持多种类型的变量。总的来说，根据变量的长度性质，Ethereum中的持久化的变量可以分为定长的变量和不定长度的变量两种。定长的变量有常见的单变量类型，比如 uint256。不定长的变量包括了由若干单变量组成的Array，以及KV形式的Map类型。

根据上面的介绍，我们了解到对 Contract Storage 层的访问是通过 Slot 的地址来进行的。请读者先思考下面的几个问题:

- **如何给定一个包含若干持久化存储变量的 Solidity 的合约，EVM 是怎么给其包含的变量分配存储空间的呢？**
- 怎么保证 Contract Storage 的一致性读写的？(怎么保证每个合约的验证者和执行者都能获取到相同的数据？)

我们将通过下面的一些实例来展示，在 Ethereum 中，Contract 是如何保存持久化变量的，以及保证所有的参与者都能一致性读写的 Contract 中的数据的。

### Contract Storage Example One

我们使用一个简单的合约来展示Contract Storage层的逻辑，合约代码如下所示。在本例中，我们使用了一个叫做"Storage"合约，其中定义了了三个持久化uint256类型的变量分别是number, number1, 以及number2。同时，我们定义一个stores函数给这个三个变量进行赋值。

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

我们使用[Remix](https://remix.ethereum.org/)来在本地部署这个合约，并构造一个调用stores(1)函数的Transaction，同时使用Remix debugger来Storage层的变化。在Transaction生效之后，合约中三个变量的值将被分别赋给1，2，3。此时，我们观察Storage层会发现，存储层增加了三个Storage Object。这三个Storage Object对应了三个Slot。所以在本例中，合约增加了三个Slots来存储数据。我们可以发现每个Storage Object由三个字段组成，分别是一个32 bytes的key字段和32 bytess的value字段，以及外层的一个32 bytes 的字段。这三个字段在下面的例子中都表现为64位的16进制数(32 Bytes)。

下面我们来逐个解释一下这个三个值的实际意义。首先我们观察内部的Key-Value对，可以发现下面三个Storage Object中key的值其实是从0开始的递增整数，分别是0，1，2。它代表了当前Slot的地址索引值，或者说该Slot在Storage层对应的绝对位置(Position)。比如，key的值为0时，它代表整个Storage层中的第1个Slot，或者说在1号位置的Slot，当key等于1时代表Storage层中的第2个Slot，以此类推。每个Storage Object中的value变量，存储了合约中三个变量的值(1,2,3)。而Storage Object外层的值由等于Storage Object的key的值的sha3的哈希值。比如，下面例子中的第一个Storage Object的外层索引值"0x290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563" 是通过keccak256(0)计算出的值，代表了第一个Slot position的Sha3的哈希，而"0xb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf6" 对应了是keccak(1)的值。我们在[示例代码](../example/account/main.go)中展示了如何计算的过程。

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

读者可能以及发现了，在这个Storage Object中，外层的索引值其实与Key值的关系是一一对应的，或者说这两个键值本质上都是关于Slot位置的唯一索引。这里我们简单讲述一下这两个值在使用上的区别。Key值代表了Slot在Storage层的Position，这个值用于会作为stateObject.go/getState()以及setState()函数的参数，用于定位Slot。如果我们继续深入上面的两个函数，我们就会发现，当内存中不存在该Slot的缓存时，geth就会尝试从更底层的数据库中来获取这个Slot的值。而Storage在更底层的数据，是由Secure Trie来维护的，Secure Trie中的Key值都是需要Hash的。所以在Secure Trie层我们查询/修改需要的键值就是外层的hash值。具体的关于Secure Trie的描述可以参考[Trie](10_tire_statedb.md)这一章节。总结下来，在上层函数(stateObject)调用中使用的键值是Slot的Position，在下层的函数(Trie)调用中使用的键值是Slot的Position的哈希值。

```go
func (t *SecureTrie) TryGet(key []byte) ([]byte, error) {
  // Secure Trie中查询的例子
  // 这里的key还是Slot的Position
  // 但是在更下层的Call更下层的函数的时候使用了这个Key的hash值作为查询使用的键值。
  return t.trie.TryGet(t.hashKey(key))
}
```

### Account Storage Example Two

下面我们来看另外的一个例子。在这个例子中，我们调整一下合约中变量的声明顺序，从(number，number1，number2)调整为(number 2, number 1, number)。合约代码如下所示。

```solidity
// SPDX-License-Identifier: GPL-3.0

pragma solidity >=0.7.0 <0.9.0;

/**
 * @title Storage
 * @dev Store & retrieve value in a variable
 */
contract Storage {

    uint256 number2;
    uint256 number1;
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

同样我们还是构造 Transaction 来调用合约中的`stores`函数。此时我们可以在 Storage 层观察到不一样的结果。我们发现number2这个变量的值被存储在了第一个 Slot 中（Key:"0x0000000000000000000000000000000000000000000000000000000000000000"），而number这个变量的值被存储在了第三个Slot中 (Key:"0x0000000000000000000000000000000000000000000000000000000000000002")。

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

这个例子可以说明，在Ethereum中，变量对应的存储层的Slot，是按照其在在合约中的声明顺序，从第一个Slot（position：0）开始分配的。

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

我们可以观察到，stores函数调用的结果只对在合约的Storage层中位置在1和2位置的两个Slot进行了赋值。值得注意的是，在本例中，对于Slot的赋值是从1号位置Slot的开始，而不是0号Slot。这说明对于固定长度的变量，其值的所占用的Slot的位置在Contract初始化开始的时候就已经分配的。即使变量只是被声明还没有真正的赋值，保存其值所需要的Slot也已经被EVM分配完毕。而不是在第一次进行变量赋值的时候，进行再对变量所需要的的Slot进行分配。

![Remix Debugger](../figs/01/remix.png)

### Account Storage Example Four

在 Solidity 中，有一类特殊的变量类型**Address**，通常用于表示账户的地址信息。例如在ERC-20合约中，用户拥有的token信息是被存储在一个(address->uint)的map结构中。在这个map中，key就是 Address 类型的，它表示了用户实际的 address 。目前Address的大小为160bits(20bytes)，并不足以填满一整个 Slot。因此当Address作为value单独存储在的时候，它并不会排他的独占用一个Slot。我们使用下面的例子来说明。

在下面的示例中，我们声明了三个变量，分别是number(uint256)，addr(address)，以及isTrue(bool)。我们知道，在以太坊中Address类型变量的长度是20 bytes，所以一个Address类型的变量是没办法填满整个的Slot(32 bytes)的。同时，布尔类型在以太坊中只需要一个bit(0 or 1)的空间. 因此，我们构造transaction并调用函数 storeaddr 来给这三个变量赋值，函数的input参数是一个uint256的值，一个address类型的值，分别为{1, “0xb6186d3a3D32232BB21E87A33a4E176853a49d12”}。

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

Transaction的运行后Storage层的结果如下面的Json所示。我们可以观察到，在本例中Contract声明了三个变量，但是在Storage层只调用了两个Slot。第一个Slot用于保存了uint256的值，而在第二个Slot中(Key:0x0000000000000000000000000000000000000000000000000000000000000001)保存了addr和isTrue的值。这里需要注意，虽然这种将两个小于32 bytes长的变量合并到一个Slot的做法节省了物理空间，但是也同样带来读写放大的问题。因为在Geth中，读操作最小的读的单位都是按照32bytes来进行的。在本例中，即使我们只需要读取isTrue或者addr这两个变量的值，在具体的函数调用中，我们仍然需要将对应的Slot先读取到内存中。同样的，如果我们想修改这两个变量的值，同样需要对整个的Slot进行重写。这无疑增加了额外的开销。所以在Ethereum使用32 bytes的变量，在某些情况下消耗的Gas反而比更小长度类型的变量要小(例如 unit8)。这也是为什么Ethereum官方也建议使用长度为32 bytes变量的原因。

// Todo Gas cost? here or in EVM Section

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

对于变长数组和Map结构的变量存储分配则相对的复杂。虽然Map本身就是key-value的结构，但是在Storage 层并不直接使用map中key的值或者key的值的sha3 哈希值来作为Storage分配的Slot的索引值。目前，Geth首先会使用map中元素的key的值和当前Map变量声明位置对应的slot的值进行拼接，再使用拼接后的值的keccak256哈希值作为Slot的位置索引(Position)。我们在下面的例子中展示了Geth是如何处理map这种变长的数据结构的。在下面的合约中，我们声明了一个定长的uint256类型的对象number，和一个[string=>uint256]类型的Map对象。

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

我们构造一个Transaction来调用`set_balance`函数。在Transaction执行之后的Storage层的结果如下面的Json所示。我们发现，对于定长的变量number占据了第一个Slot的空间(Position:0x0000000000000000000000000000000000000000000000000000000000000000)。但是对于Map类型变量balances，它包含的两个数据并没有按照变量定义的物理顺序来定义Slot。此外，我们观察到存储这两个值的Slot的key，也并不是这两个字在mapping中key的直接hash。正如我们在上段中提到的那样，Geth会使用Map中元素的的key值与当前Map被分配的slot的位置进行拼接，之后对拼接之后对值进行使用keccak256函数求得哈希值，来最终得到map中元素最终的存储位置。比如在本例中，按照变量定义的顺序，balances这个Map变量会被分配到第二个Slot，对应的Slot Position是1。因此，balances中的kv对分配到的slot的位置就是，keccak(key, 1)，这里是一个特殊的拼接操作。

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

为了验证上面的说法，我们使用 go 语言编写了一段代码，来调用相关的库来验证一下上面的结论。对于 balances["hsy"]，它被分配的Slot的位置可以由下面的代码求得。读者可以阅读/使用[示例代码](../example/account/main.go)进行尝试。这里的k1是一个整形实数，代表了Slot的在storage层的位置(Position)。

```go
k1 := solsha3.SoliditySHA3([]byte("hsy"), solsha3.Uint256(big.NewInt(int64(1))))
fmt.Printf("Test the Solidity Map storage Key1:         0x%x\n", k1)
```

// TODO: The wallet part. Not sure in this section or another section.

## Wallet

- KeyStore
- Private Key
- 助记词

## Reference

- <https://www.freecodecamp.org/news/how-to-generate-your-very-own-bitcoin-private-key-7ad0f4936e6c/>
