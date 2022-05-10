# EVM in Practice

## EVM Instructions

关于EVM的指令，我们首先关注会与StateDB交互，甚至会引发Disk I/O的指令。这些指令包括`Balance`,`Sload`,`Sstore`, `EXTCODESIZE`,`EXTCODECOPY`,`EXTCODEHASH`,`SELFDESTRUCT`,`LOG0`,`LOG1`,`LOG2`,`LOG3`,`LOG4`,`KECCAK256`。其中`LOG0`,`LOG1`,`LOG2`,`LOG3`,`LOG4`在本质上都调用了同一个底层函数`makeLog`。

## EVM Trace

### EVM Logger