# 从开发的角度探究go-ethereum的设计思想

go-ethereum 中大量使用了新的实例来处理业务逻辑。比如在core/state_processor.go中:

```go
func (p *StateProcessor) Process(block *types.Block, statedb *state.StateDB, cfg vm.Config) (types.Receipts, []*types.Log, uint64, error) {
    ....
    blockContext := NewEVMBlockContext(header, p.bc, nil)
    // 新new了一个EVM的实例
	vmenv := vm.NewEVM(blockContext, vm.TxContext{}, statedb, p.config, cfg)
    ....
}

```

Go-ethereum很多函数将配置文件/协议作为函数的参数，来动态的修改/升级系统的逻辑。比如:
/core/state_transition.go/refundGas()用于当transaction执行失败时返还给用户的gas，其参数是最高返还的比例(份额)。
```go
func (st *StateTransition) refundGas(refundQuotient uint64) {
    // Apply refund counter, capped to a refund quotient
    refund := st.gasUsed() / refundQuotient
    if refund > st.state.GetRefund() {
        refund = st.state.GetRefund()
    }
    st.gas += refund

    // Return ETH for remaining gas, exchanged at the original rate.
    remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)
    st.state.AddBalance(st.msg.From(), remaining)

    // Also return remaining gas to the block gas counter so it is
    // available for the next transaction.
    st.gp.AddGas(st.gas)
}
```

这个函数的调用在/core/state_transition.go/TransitionDb()调用的时候，输入是根据EIP协议的动态配置params.RefundQuotientEIP3529.

```go
func (st *StateTransition) TransitionDb() (*ExecutionResult, error) {
    ....
    if !st.evm.ChainConfig().IsLondon(st.evm.Context.BlockNumber) {
            // Before EIP-3529: refunds were capped to gasUsed / 2
        st.refundGas(params.RefundQuotient)} 
        else {
        // After EIP-3529: refunds are capped to gasUsed / 5
        st.refundGas(params.RefundQuotientEIP3529)
    }
    ....
}
```