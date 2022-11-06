# EVM in Practice

## General

EVM是Ethereum中最核心的模块，也可以说是Ethereum运行机制中的灵魂模块。

如果我们抛开Blockchain的属性，那么EVM的JVM(或者其他的类似的Virtual Machine)。EVM更类似于一个解释器(Interpreter)而不是传统的编译器(Complier)。

EVM，由Stack，Program Counter，Gas available，Memory，Storage，Opcodes组成。

![The EVM Workflow](../figs/14/EVM%20Flow.png)

## EVM Instructions

关于EVM的指令，我们首先关注会与StateDB交互，甚至会引发Disk I/O的指令。这些指令包括`Balance`,`Sload`,`Sstore`, `EXTCODESIZE`,`EXTCODECOPY`,`EXTCODEHASH`,`SELFDESTRUCT`,`LOG0`,`LOG1`,`LOG2`,`LOG3`,`LOG4`,`KECCAK256`。其中`LOG0`,`LOG1`,`LOG2`,`LOG3`,`LOG4`在本质上都调用了同一个底层函数`makeLog`。

## EVM Trace

我们知道，在以太坊中有两种类型的交易，1. Native 的Ether的转账交易 2. 调用合约函数的交易。调用合约的交易，本质上是实行了一段函数代码，由于是图灵完备的，这算代码可以任意的运行。作为用户，我们只需要知道Transaction的最终的运行结果(最终修改的Storage的结果)。但是对于开发人员，我们需要了解交易运行的最终结果，我们还需要了解在Transaction执行过程中的一些中间状态来方便debug和调优。

为了满足开发人员的这种需求，go-ethereum提供了EVM Tracing的模块，来trace Transaction执行时EVM中一些值的情况。

目前，这种EVM的Trace是以transaction执行时的调用的opcode单位开展的。EVM每执行一个指令，都会将当前Stack，Memory，Storage，Transaction剩余的Gas量，当前执行的指令的信息输出出来。

### EVM Logger


## Reference

- [An Ethereum Virtual Machine
Opcodes Interactive Reference](https://www.evm.codes/)