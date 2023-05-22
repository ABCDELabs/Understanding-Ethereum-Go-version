# Mining：挖矿

## Block Reward：区块奖励

## How to Seal Block：结算区块

其中一个关键的函数是`miner/worker.go`中的`fillTransactions()`函数。这个函数会从transaction pool中的Pending的transactions选取若干的交易并且将他们按照Gas Price和Nonce的顺序进行排序形成新的tx set并传递给`commitTransactions()`函数。在`fillTransactions()`函数会首先处理Local Pool中的交易，然后再处理从网络中接受到的远程交易。

目前Block中的Transaction的顺序就是由`fillTransactions()`函数通过调用`types/transaction.go`中的`NewTransactionsByPriceAndNonce()`来决定。

也就说如果我们希望修改Block中Transaction的打包顺序和从Transaction Pool选择Transactions的策略的话，我们可以通修改`fillTransactions()`函数。

`commitTransactions()`函数的主体是一个for循环体。在这个for循环中，函数会从txs中不断拿出头部的tx进行调用`commitTransaction()`函数进行处理。在Transaction那一个Section我们提到的`commitTransaction()`函数会将成功执行的Transaction保存在`env.txs`中。

