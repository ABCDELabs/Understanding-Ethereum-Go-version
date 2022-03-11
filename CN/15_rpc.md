# Ethereum中的API 调用: RPC and IPC

## RPC (远程过程调用)

### Background

一个应用可以通过RPC(Remote Procedure Call)的方式来调用某一个go-ethereum的实例(instance)。 通常，go-ethereum默认的对外暴露的RPC端口地址为8545。

在example/deploy/SendTransaction.go中，我们展示了一个通过RPC调用go-ethereum实例来发送Transaction的例子。

## IPC (进程间通信)

### Background

IPC(Inter-Process Communication) 用于两个进程间进行通信，和共享数据。与RPC不同的是，IPC主要用在同一台宿主机上的不同的进程间的通信。

## Appendix
