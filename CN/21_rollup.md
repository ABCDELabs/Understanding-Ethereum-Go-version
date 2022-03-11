# 扩容！扩容！扩容！从Plasma到Rollup

## In General

## 一些旧日的想法: Plasma

## 弄潮儿Rollup

自从[EIP-1559](https://eips.ethereum.org/EIPS/eip-1559)生效后，由于币价的升高，和基础Gas Price的约束，Ethereum Mainnet上的单笔交易费用已经高到了离谱的程度。显然，目前**天价的交易费**以及**有限的Throughput**已经成为了限制Ethereum继续发展的两大难题。幸运的是，**rollup**技术的发展给社区展现了一种似乎可以一招解决两大难题的绝世武学。

简单的来说，顾名思义，rollup，就是把一堆的transaction rollup到一个新的transaction。然后，通过某种神奇的技术，使得Ethereum Mainnet只需要验证这个新生成的transaction，就可以保证被Rollup之前的若干的Transaction的确定性，正确性，完整性。举个简单的例子，我们想象一个学校内交学费的场景。过去，每个学生(i.e. Account)都需要通过学校交费系统(i.e. Ethereum Mainnet)单独的将自己的学费转账(Transfer)给学校教务处。假如，现在计算机系有1000个学生，那么系统就要处理1000笔转账的交易(Transaction)。现在，系里新来一位叫做*Rollup*的教授。他在学生中很有号召力。基于他的个人魅力(神奇的魔法)，**私下里**让所有的的学生把钱通过**某种方式**先转给他。当*Prof.Rollup*收集到系里所有学生的学费之后，然后他通过构造一个transaction把所有的学费交给教务处。这个例子大概就是一个Rollup场景的简单抽象。我们把学校的交费系统看作为Ethereum Mainnet 或者说是Layer-1，那么私下里收集学生的学费就是所谓的**Layer-2**进行的了。通常，我们把Layer-1上的交易称为on-chain transaction，把在Layer-2进行上的交易称为off-chain transaction。

Rollup的好处是显而易见的，假如每次把**N**个transaction rollup成一个，那么对于Mainnet来说，处理一条交易，实际效果等同于处理了之前系统中的**N**个交易。同时，系统实际吞吐量的Upper Bound也实际上上市到了**N*MAX_TPS**的水平。同时，对于用户来说，同样一条交易Mainnet Transaction Fee实际上是被**N**个用户同时负担的。那么理论上，用户需要的实际交易费用也只有之前的**1/N+c**。这里的**c**代表rollup服务提供商收取的交易费，这个值是远远小于layer-1上交易所需要的交易费用的。

Rollup看上去完美的解决了Ethereum面临的两大难题，通过分层的方式给Ethereum进行了扩容。但是，在实现layer-2 rollup时，还是有很多的细节有待商榷。比如:

- 怎么保证*Prof.Rollup*一定会把学费交给教务处呢？ (Layer-2 交易的安全性)
- 怎么保证*Prof.Rollup*会把全部的学费都交给教务处呢？(Layer-2 交易的完整性)
- *Prof.Rollup*什么时候才会把学费打给教务处呢？(Layer-2 到Layer-1 跨链交易的时效性问题)
- 等等..

就像华山派的剑宗和气宗分家一样，目前，在如何实现Rollup上，主要分为了两大流派，分别是**Optimism-Rollup**，和**ZK-Rollup**。两种路线各有所长，又各有不足。

## 弄潮儿的精英Zk-Rollup

ZK-Rollup核心是利用了Zero-Knowledge Proof的技术来实现rollup。ZKP里面的细节比较多，在这里我们不展开描述。其中最重要的核心点是ZK-Rollup主要利用了ZK-SNARKs中的可验证性的正确性保障，以及Verification Time相对较快的特性，来保证layer-1上的Miner可以很快的验证大量交易的准确性。

但是正因为ZKP的一些特性，使得ZK-Rollup相比于Optimism-Rollup，在开发上的并没有进行的那么顺利。

我们知道ZK-SNARKs的计算的基础来自于:*将一个计算电路(Circuit)，转化为R1CS的形式*，继而转化为QAP问题，最终将问题的Witness生成零知识的Proof。如果我们想生成一个Problem/Computation/Function的ZK-SNARKs的Witness/Proof，那么首先我们需要要把这个问题转化为一个Circuit的形式。或者说用Circuit的语言，用R1CS的形式来描述原计算问题。对于简单的计算问题来说，比如加减计算，解方程组，将问题转化为电路的形式并不是特别的困难。但是对于Ethereum上各种支持图灵完备的智能合约的function来说这个问题就变得非常的棘手。主要因为：

- 转化生成的电路是静态的，虽然电路中有可以复用的部分(Garget)，但是每个Problem/Computation/Function都要构造新的电路。
- 对于通用计算问题，构造出来的电路需要的gate数量可能是惊人的高。这意味着ZK-Rollup可能需要非常长的时间来生成Proof。
- 对于目前EVM架构下的某些计算问题，生成其电路是非常困难的。

前两个问题属于General的ZK-SNARKs应用都会遇到的问题，目前已经有一些的研究人员/公司正在尝试解决这个问题。比如AleoHQ的创始人Howard Wu提出的[DIZK](https://www.usenix.org/conference/usenixsecurity18/presentation/wu)通过分布式的Cluster来并发计算Proof，以及Scroll的创始人[Zhang Ye](https://twitter.com/yezhang1998)，提出的[PipeZK](https://www.microsoft.com/en-us/research/publication/pipezk-accelerating-zero-knowledge-proof-with-a-pipelined-architecture/)通过ASIC来加速ZK计算。

对于第三个问题，属于Ethereum中的专有的问题，也是目前zk-Rollup实现中最难，最需要攻克的问题。我们知道在Ethereum中，一条调用合约Function的Transaction的执行是基于/通过EVM的来执行的。EVM会将Transaction中的函数调用，基于合约代码，转换成opcodes的形式，保存在Stack中逐条执行。这个过程类似于编译过程中的IR代码生成，或者高级语言到汇编语言的过程。在执行这些opcodes时，本质上是在执行geth中对应的库函数，部分细节可以参考之前的[blog](http://www.hsyodyssey.com/blockchain/2021/07/25/ethereum_txn.html)。那么如果我们想把一个Transaction的合约调用转换成一个电路的话，本质上我们就要基于这个函数调用过程中执行过的opcodes来生成电路。目前的问题是，EVM在设计的时候并没有考虑到将来会被ZK-SNARKs这一问题，所以在它包含的140个opcode中有些是难以构造电路的。

结果就是，在目前的ZK-Rollup的解决方案中，大部分**仅支持基础的转账操作**，而**不能支持**通用的图灵完备的计算。也就是说，目前的ZK-Rollup的解决方案都是不完整的，不能完整的发挥Ethereum图灵完备合约的特性，Layer-2又回到了仅支持Token转账的时代。

为了解决这个问题，使得Layer-2能支持像现在的Layer-1一样的功能，目前的技术主要在朝向两个方向发展，1. 构建ZK-SNARKs兼容的zkEVM，2.提出新的VM来兼容ZK-SNARKs，并想办法与Ethereum兼容。
