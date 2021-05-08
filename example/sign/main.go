package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
)

var AlicePrivateKey = "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"

func main() {
	dataHash := sha256.Sum256([]byte("ethereum"))

	// The account's Private Key
	privateKey, err := hex.DecodeString(AlicePrivateKey)

	if err != nil {
		log.Fatalln("PK Generate Failed", err)
	}

	// Generate PK
	sk, err := crypto.ToECDSA(privateKey)
	if err != nil {
		log.Fatalln("SK Generate Failed", err)
	}

	fmt.Println("Account Address is:     ", crypto.PubkeyToAddress(sk.PublicKey))
	fmt.Println("Account Public key is:  ", hex.EncodeToString(crypto.FromECDSAPub(&sk.PublicKey)))
	fmt.Println("Account Private key is: ", hex.EncodeToString(privateKey))

	// Sign the Data
	sig, err := crypto.Sign(dataHash[:], sk)
	if err != nil {
		log.Fatalln("Sign Failed", err)
	}

	// fmt.Println("sig len:                ", len(sig))
	// fmt.Println("sig hex:                ", hex.EncodeToString(sig))

	decodeHex := func(s string) []byte {
		b, err := hex.DecodeString(s)
		if err != nil {
			log.Fatal(err)
		}
		return b
	}

	sigTest := decodeHex(hex.EncodeToString(sig))

	// Recover the Public Key of the Account from the Sig and the Message
	recoveredPub, err := crypto.Ecrecover(dataHash[:], sigTest)
	if err != nil {
		log.Fatal(err)
	}

	// Get the Account Public Key
	pubKey, _ := crypto.UnmarshalPubkey(recoveredPub)

	// Get the Account Public Key Bytes
	recoveredPubBytes := crypto.FromECDSAPub(pubKey)

	fmt.Println("Recover the Public Key: ", recoveredPub)
	fmt.Println("UnmarshalPubkey:        ", pubKey)
	fmt.Println("Account Pubkey bytes:   ", recoveredPubBytes)
	fmt.Println("Account Public Key:     ", hex.EncodeToString(recoveredPubBytes))
	fmt.Println("Compress the Public Key: ", crypto.CompressPubkey(pubKey))

	fmt.Println("----------------------------------------------")

	testPk := decodeHex("037db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf7")

	// Verify by the original public key
	longVerify := crypto.VerifySignature(recoveredPub, dataHash[:], sigTest[:len(sigTest)-1])

	fmt.Println("[Original Public Key] verify pass?", longVerify)

	// Verify by the compressed public key
	shortVerify := crypto.VerifySignature(testPk, dataHash[:], sigTest[:len(sigTest)-1])

	fmt.Println("[Compressed Public Key] verify pass?", shortVerify)

}
