# Account

## 基本数据结构
~~在之前的版本中Account的代码位于core/account.go~~
在最新的Go-Ethereum 版本中Account 被抽象成了State_object,代码位于core/state/state_object.go

```go
// Account is the Ethereum consensus representation of accounts.
// These objects are stored in the main account trie.
type Account struct {
  Nonce    uint64
  Balance  *big.Int
  Root     common.Hash // merkle root of the storage trie
  CodeHash []byte
}
```

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
  data     Account
  db       *StateDB
  dbErr error

  // Write caches.
  trie Trie // storage trie, which becomes non-nil on first access
  code Code // contract bytecode, which gets set when code is loaded

  originStorage  Storage // Storage cache of original entries to dedup rewrites, reset for every transaction
  pendingStorage Storage // Storage entries that need to be flushed to disk, at the end of an entire block
  ....
}
```

## Account & Private Key & Public Kay & Address

- 首先我们通过随机得到一个长度64位account的私钥。
  - 64个16进制位，256bit，32字节
    `var AlicePrivateKey = "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"`

- 得到私钥后我们用私钥来计算公钥和account的地址。对基于私钥，我们使用ECDSA算法，选择spec256k1曲线，进行计算。通过将私钥带入到所选择的椭圆曲线中，计算出点的坐标即是公钥。
以太坊和比特币使用了同样的spec256k1曲线，在实际的代码中，我们也可以看到在crypto中，go-Ethereum调用了比特币的代码。

    `ecdsaKey, err := crypto.ToECDSA(privateKey)`

- 对私钥进行椭圆加密之后，我们可以得到64字节的数，是由两个32字节的数构成，这两个数代表了spec256k1曲线上某个点的XY值。
    `pk := ecdsaKey.PublicKey`
- 以太坊的地址，是基于上述公钥的Keccak-256算法之后的后20个字节，并且用0x开头。
  - Keccak-256是SHA-3（Secure Hash Algorithm 3）标准下的一种哈希算法
    `addr := crypto.PubkeyToAddress(pk.PublicKey)`

## Signature & Verification

- Hash（m,R）*X +R = S * P
- P是椭圆曲线函数的基点(base point) 可以理解为一个P是一个在曲线C上的一个order 为n的加法循环群的生成元. n为质数。
- R = r * P (r 是个随机数，并不告知verifier)
- 以太坊签名校验的核心思想是基于上面得到的ecdsaKey对数据msg进行签名得到msgSig. 
    `sig, err := crypto.Sign(msg, ecdsaKey)`
- 基于msg和msgSig可以反推出来签名的公钥（生成地址的那个）。
    `recoveredPub, err := crypto.Ecrecover(dataHash[:], sigTest)`
- 通过反推出来的公钥得到发送者的地址，并与当前txn的发送者进行对比。
    `crypto.VerifySignature(testPk, msg, msgSig)`
- 这套体系的安全性保证在于，即使知道了公钥pk/ecdsaKey.PublicKey也难以推测出 ecdsaKey以及生成他的privateKey。

## ECDSA & spec256k1曲线

- Elliptic curve point multiplication
  - Point addition P + Q = R
  - Point doubling P + P = 2P
- y^2 = x^3 +7
- Based Point P是在椭圆曲线上的群的生成元
- x次computation on Based Point得到X点，x为私钥，X为公钥。x由Account Private Key得出。
- 在ECC中的+号不是四则运算中的加法，而是定义椭圆曲线C上的新的二元运算(Point Multiplication)。他代表了过两点P和Q的直线与椭圆曲线C的交点R‘关于X轴对称的点R。因为C是关于X轴对称的所以关于X对称的点也都在椭圆曲线上。

## Code Example

这部分的示例代码在[example/account](example/account)中。
