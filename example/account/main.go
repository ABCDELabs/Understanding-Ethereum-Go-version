package main

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/crypto"

	solsha3 "github.com/miguelmota/go-solidity-sha3"
)

var toHash = common.BytesToHash

func main() {
	var snaps *snapshot.Tree
	stateDB, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), snaps)

	account1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	account2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	stateDB.AddBalance(account1, big.NewInt(2000))
	stateDB.AddBalance(account2, big.NewInt(8888))
	contract := crypto.CreateAddress(account1, stateDB.GetNonce(account1))
	stateDB.CreateAccount(contract)
	stateDB.SetCode(contract, []byte("contract code bytes"))

	stateDB.SetNonce(contract, 1)
	stateDB.SetState(contract, toHash([]byte("owner")), toHash(account1.Bytes()))
	stateDB.SetState(contract, toHash([]byte("name")), toHash([]byte("hsy")))

	stateDB.SetState(contract, toHash([]byte("online")), toHash([]byte{1}))
	stateDB.SetState(contract, toHash([]byte("online")), toHash([]byte{}))

	stateDB.Commit(true)

	fmt.Println(string(stateDB.Dump(true, true, true)))

	fmt.Println("------Test Hash-------")

	for i := 0; i <= 2; i++ {
		hash := solsha3.SoliditySHA3(
			solsha3.Uint256(big.NewInt(int64(i))),
		)
		fmt.Printf("The hash of %d:   %x\n", i, hash)
	}

}
