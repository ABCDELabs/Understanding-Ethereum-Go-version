# 一个Transaction的生老病死/Transaction CRUD


## State-based Blockchain

State-based Blockchain System 由两部分组成：World State 和 Blockchain。

Blockchain是以块为单位的数据结构，每个块中包含了若干Transaction。

World State表示了System中所有Account的值的一个Snapshot。

这个Snapshot的建立是以Block为单位的。

## Transaction是如何被打包并修改Blockchain中的值的

Transaction是用于修改Account的State的。若干个Transaction对修改的结果整合成一个新的World State


当miner开始构造新的区块的时候。首先调用 miner/worker.go的 mainLoop() 函数。

    ```go
    func (w *worker) mainLoop() {
        ....
        coinbase := w.coinbase
        w.mu.RUnlock()

        txs := make(map[common.Address]types.Transactions)
        for _, tx := range ev.Txs {
            acc, _ := types.Sender(w.current.signer, tx)
            txs[acc] = append(txs[acc], tx)
        }
        txset := types.NewTransactionsByPriceAndNonce(w.current.signer, txs, w.current.header.BaseFee)
        tcount := w.current.tcount
        w.commitTransactions(txset, coinbase, nil)        
        ....
    }
    ```

Worker会从TransactionPool中拿出若干的transaction, 赋值给*txs*, 然后按照Price和Nonce对*txs*进行排序，并将结果赋值给*txset*。

在拿到*txset*之后，mainLoop函数会调用miner/worker.go的commitTransactions()函数。

    ```go
    func (w *worker) commitTransactions(txs *types.TransactionsByPriceAndNonce, coinbase common.Address, interrupt *int32) bool {
        ....

        // 首先给Block设置最大可以使用的Gas的上限
        gasLimit := w.current.header.GasLimit
        if w.current.gasPool == nil {
        w.current.gasPool = new(core.GasPool).AddGas(gasLimit)
        // 函数的主体是一个For循环
        for{
        .....
            // params.TxGas表示了transaction 需要的最少的Gas的数量
            // w.current.gasPool.Gas()可以获取当前block剩余可以用的Gas的Quota，如果剩余的Gas足以开启一个新的Tx，那么循环结束
            if w.current.gasPool.Gas() < params.TxGas {
                log.Trace("Not enough gas for further transactions", "have", w.current.gasPool, "want", params.TxGas)break
        }
        ....
        logs, err := w.commitTransaction(tx, coinbase)
        ....
        }
    }
}
    ```
对于每一个tx in *txs*，调用函数miner/worker.go的commitTransaction()

    ```go

    ```


    ```go

    ```


Worker get from transaction pool.

+ commitTransactions
  + commitTransaction ->> ApplyTransaction ->> ApplyMessage ->> TransactionDB

## Reference

+ https://www.codenong.com/cs105936343/
+ https://yangzhe.me/2019/08/12/ethereum-evm/