# Ethereum Transaction Pool

Transaction pool也称为Mempool是节点用于临时存储尚未被打包，处于待打包状态的transaction。


当节点新收到来自网络的transaction 之后，首先transaction pool会检查其合法性。代码位于core/tx_pool.go。

    func (pool *TxPool) validateTx(tx *types.Transaction, local bool) error

