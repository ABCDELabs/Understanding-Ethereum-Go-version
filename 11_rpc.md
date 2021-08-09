# RPC (远程过程调用)

## Background

一个应用可以通过RPC(Remote Procedure Call)的方式来调用某一个go-ethereum的实例(instance)。 通常，go-ethereum默认的对外暴露的RPC端口地址为8545。

## Appendix

在example/deploy/SendTransaction.go中，我们展示了一个通过RPC调用go-ethereum实例来发送Transaction的例子。