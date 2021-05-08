package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func SendTx() {

	account := common.HexToAddress("You Account")

	data, _ := ioutil.ReadFile("key")
	fmt.Println(data)
	fmt.Println(string(data))

	client, err := ethclient.Dial("The RPC interface address")
	if err != nil {
		log.Fatal("Unable to connect to network:", err)
	}

	// Get the Balance
	balance, err := client.BalanceAt(context.Background(), account, nil)

	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	bnbValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))

	fmt.Println(bnbValue)

	if err != nil {
		log.Fatal(err)
	}

	privateKey, err := crypto.HexToECDSA("You private Key")
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()

	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)

	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	// 18 zeros, transfer 1 ETH
	value := big.NewInt(1000000000000000000)

	gasLimit := uint64(21000)

	// Get the estimate Gas Price (Based on the average gas price in previous transaction gas price)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Construct the target Address
	toAddress := common.HexToAddress("Target Address")

	// Construct the raw Transaction (Unsigned)
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	// Get the current ChainID
	ChainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Sign the transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(ChainID), privateKey)

	if err != nil {
		log.Fatal(err)
	}

	// Send Signed Transactions
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

}
