# Ethereum Transaction Pool

Transaction pool也称为Mempool是节点用于临时存储尚未被打包，处于待打包状态的transaction。


当节点新收到来自网络的transaction 之后，首先transaction pool会检查其合法性。代码位于core/tx_pool.go。

    func (pool *TxPool) validateTx(tx *types.Transaction, local bool) error

当Newly added transaction确定是合法交易之后，transaction 会被首先添加到一个non-executable queue中缓存。在之后的某一时刻会被pending promotion and execution。
这里有两个地方要注意：
    - 新添加且验证合法的交易并不会马上被节点添加到Pending list进入待打包状态。而是会被送到一个non-executable queue的队列中缓存。
        func (pool *TxPool) enqueueTx(hash common.Hash, tx *types.Transaction, local bool, addAll bool) (bool, error)
    - 新添加且验证合法的交易在被打包之前并不会此刻（在tx pool中等待的时刻）被执行。