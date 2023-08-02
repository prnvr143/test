package accounts

import (
	"github.com/tyler-smith/go-bip39"
)

var Menmonic string

func GenerateMemonic() string {
	entrphy, _ := bip39.NewEntropy(256)
	////fmt.Println("entrphy ->", entrphy)
	Menmonic, _ := bip39.NewMnemonic(entrphy)
	////fmt.Println("mnemonic ->", Menmonic)
	return Menmonic
}

func GenerateSeed(mnemonic string) []byte {
	seed := bip39.NewSeed(mnemonic, "password from user") // later on change this to take a input using io.Reader(maaybe)
	////fmt.Println("Your seed is ->", seed)
	return seed
}
