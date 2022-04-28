package main

import (
	"fmt"
	"time"
)

type Node struct {
	backend Backend
	stop    chan struct{}
}

type Backend struct {
	balance int
}

func (n *Node) Wait() {
	<-n.stop
}

func (b *Backend) Mainloop() {
	for ; b.balance != 0; b.balance-- {
		time.Sleep(1 * time.Second)
		fmt.Println("Current Balance: ", b.balance)
	}
}

func NewNode() *Node {
	n := &Node{
		backend: *NewBackend(),
		stop:    make(chan struct{}),
	}
	return n
}

func NewBackend() *Backend {
	b := &Backend{
		balance: 100,
	}
	return b
}

func main() {
	n := NewNode()
	go n.backend.Mainloop()
	go func() {
		// Close the Node after 10 second
		time.Sleep(10 * time.Second)
		close(n.stop)
	}()
	n.Wait()
}
