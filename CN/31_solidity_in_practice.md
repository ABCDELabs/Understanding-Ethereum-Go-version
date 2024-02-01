# Solidity

## Instructions

EVM类似汇编器，负责把合约汇编成更底层的指令(instruction)。每条指令表示了一些基础或者原子行逻辑操作，例如opCreate用于在State Database上创建一个新的Contract，opBalance用于从State Database中获取某个State Object的balance。这些指令的的具体的代码实现位于core/vm/instructions.go 文件中。

值得注意的是，这些指令仍然会调用go-ethereum中其他package所提供的API，而不是直接对更底层的数据进行操作。比如，opSstore与opSload指令用于从Storage层存储和读取数据。这两个指令直接调用了StateDB(core/state/statedb.go)与StateObject(core/state/state_object.go)提供的API。关于这些指令的详细介绍可以参考Ethereum Yellow Paper，具体实践可以参考https://www.evm.codes/

### opSload（将栈顶位置的数据作为KEY，返回当前合约相应位置的数据）

opSload的代码如下所示。

```Golang
func opSload(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
    loc := scope.Stack.peek() //读取栈顶位置的值，这里使用.peek而不是.pop的原因是因为sload具有返回值，如果使用pop，还需要在后续再push返回的结果。
    hash := common.Hash(loc.Bytes32())
    // 从StateDB中读取到对应的合约中对应的存储Object的值
    val := interpreter.evm.StateDB.GetState(scope.Contract.Address(), hash)
    loc.SetBytes(val.Bytes())
    return nil, nil
}
```

### opSstore（将栈顶、次栈顶位置作为存储值、存储位置，传入到当前合约的DB当中）

opSstore的代码如下所示。

```Golang
func opSstore(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
    loc := scope.Stack.pop() //sstore无返回值，直接使用.pop
    val := scope.Stack.pop()
    // 将Stack中的数据写入到StateDB中
    interpreter.evm.StateDB.SetState(scope.Contract.Address(),
    loc.Bytes32(), val.Bytes32())
    return nil, nil
}
```

我们注意到，opSstore指令中向合约中的写入逻辑是调用了StateDB中的SetState函数(在core/state/statedb.go中)。SetState函数有三个参数作为input，分别是目标合约的地址，目标storage object的has，以及其更新后的value。其代码如下所示。

```Golang
func (s *StateDB) SetState(addr common.Address, key, value common.Hash) {
    stateObject := s.GetOrNewStateObject(addr)
    if stateObject != nil {
        stateObject.SetState(s.db, key, value)
        }
}
```

SetState 函数通过调用StateObject的SetState函数来修改Storage的值。

```Golang
// SetState updates a value in account storage.
func (s *stateObject) SetState(db Database, key, value common.Hash) {
    // If the fake storage is set, put the temporary state update here.
    if s.fakeStorage != nil {
        s.fakeStorage[key] = value
        return
    }
    // If the new value is the same as old, don't set
    prev := s.GetState(db, key)
    if prev == value {
        return
    }
    // New value is different, update and journal the change
    s.db.journal.append(storageChange{
        account:  &s.address,
        key:      key,
        prevalue: prev,
    })
    s.setState(key, value)
}

func (s *stateObject) setState(key, value common.Hash) {
    s.dirtyStorage[key] = value
}
```

这里的dirtStorage起到了一个cache的作用。之后在updated storage root的时候会基于当前dirtyStorage中的信息，在commit函数中统一更新root的值。
