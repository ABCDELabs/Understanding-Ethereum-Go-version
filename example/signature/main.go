package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
)

// 随机产生的一个私钥
var AlicePrivateKey = "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"

func main() {
	// 需要加密的数据
	msg := sha256.Sum256([]byte("ethereum"))

	// The account's Private Key
	privateKey, err := hex.DecodeString(AlicePrivateKey)

	if err != nil {
		log.Fatalln("PK Generate Failed", err)
	}

	// Generate SK，通过privatekey来生成ECDSA中的对应的私钥SK
	ecdsaSK, err := crypto.ToECDSA(privateKey)
	if err != nil {
		log.Fatalln("SK Generate Failed", err)
	}
	// 通过ECDSA下的pk来生成Account Address
	fmt.Println("Account Address is:     ", crypto.PubkeyToAddress(ecdsaSK.PublicKey))
	fmt.Println("Account Private key is: ", hex.EncodeToString(privateKey))
	fmt.Println("-----------------Original-----------------------")
	fmt.Println("[Original ECDSA] Account Public key is:  ", hex.EncodeToString(crypto.FromECDSAPub(&ecdsaSK.PublicKey)))
	fmt.Println("[Original ECDSA] Account Public key is:  ", hex.EncodeToString(crypto.CompressPubkey(&ecdsaSK.PublicKey)))

	// Sign the Data
	sig, err := crypto.Sign(msg[:], ecdsaSK)
	if err != nil {
		log.Fatalln("Sign Failed", err)
	}

	decodeHex := func(s string) []byte {
		b, err := hex.DecodeString(s)
		if err != nil {
			log.Fatal(err)
		}
		return b
	}

	// 签名后的数据
	msgSig := decodeHex(hex.EncodeToString(sig))

	// Recover the Public Key of the Account from the Sig and the Message
	// 通过transaction的原文和签名后的数据来恢复ECDSA下对应的Pk
	recoveredPub, err := crypto.Ecrecover(msg[:], msgSig)
	if err != nil {
		log.Fatal(err)
	}

	// Get the Account Public Key
	pubKey, _ := crypto.UnmarshalPubkey(recoveredPub)

	// Get the Account Public Key Bytes
	recoveredPubBytes := crypto.FromECDSAPub(pubKey)
	fmt.Println("------------------Recovered--------------------")

	fmt.Println("[Recovered] Account Public Key:     ", hex.EncodeToString(recoveredPubBytes))
	fmt.Println("[Recovered] Account Compressed Public Key: ", hex.EncodeToString(crypto.CompressPubkey(pubKey)))
	// fmt.Println("[Recovered] Recover the Public Key: ", recoveredPub)
	// fmt.Println("[Recovered] UnmarshalPubkey:        ", pubKey)
	// fmt.Println("[Recovered] Account Pubkey bytes:   ", recoveredPubBytes)

	fmt.Println("----------------------------------------------")

	testPk := decodeHex("037db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf7")

	// Verify by the original public key
	longVerify := crypto.VerifySignature(recoveredPub, msg[:], msgSig[:len(msgSig)-1])

	fmt.Println("[Original Public Key] verify pass?", longVerify)

	// Verify by the compressed public key
	shortVerify := crypto.VerifySignature(testPk, msg[:], msgSig[:len(msgSig)-1])

	fmt.Println("[Compressed Public Key] verify pass?", shortVerify)

}
