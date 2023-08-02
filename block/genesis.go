package block

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"

	"jumbochain.org/enum"
	db "jumbochain.org/ldb"
	"jumbochain.org/transaction"
)

type AllocJosn struct {
	Allocs []Alloc `json:"alloc"`
}

type Alloc struct {
	Address string `json:"address"`
	Balance int    `json:"balance"`
}

func InitGenesis() {

	var allocJosn AllocJosn = readGenesis("genesis.json")
	var transaction_hashes []string
	for i := 0; i < len(allocJosn.Allocs); i++ {

		to := allocJosn.Allocs[i].Address
		value := allocJosn.Allocs[i].Balance
		hash, trxBody := transaction.SignGenesisTrx(to, value)

		//add Transaction
		db.AddDataToDatabase("database", []byte(hash), []byte(trxBody))

		//update blockstate
		db.AddDataToDatabase("database", []byte(to), []byte(strconv.Itoa(value)))

		transaction_hashes = append(transaction_hashes, hash)

	}

	hash, blockBody := GenesisBlockHash(transaction_hashes)

	//update blockchain
	db.AddDataToDatabase("database", []byte(hash), []byte(blockBody))

	blockNumber := "0"
	//update block number
	db.AddDataToDatabase("database", []byte(blockNumber), []byte(hash))

	lastHash := "lh"
	//update lasthash
	db.AddDataToDatabase("database", []byte(lastHash), []byte(hash))

	//update current block number
	currentBlockNumber := "currentBlockNumber"

	db.AddDataToDatabase("database", []byte(currentBlockNumber), []byte(blockNumber))

	//update current block extended number
	currentBlockExtendedNumber := string(enum.CurrentBlockExtendedNumber)

	db.AddDataToDatabase(string(enum.Blockextended), []byte(currentBlockExtendedNumber), []byte(blockNumber))

	executedBlockExtendednumber := string(enum.ExecutedBlockExtendednumber)

	db.AddDataToDatabase(string(enum.Blockextended), []byte(executedBlockExtendednumber), []byte(blockNumber))

}

func readGenesis(filename string) AllocJosn {
	jsonFile, err := os.Open(filename)

	if err != nil {
		////fmt.Println("error is opening file")
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var allocJosn AllocJosn

	json.Unmarshal(byteValue, &allocJosn)

	return allocJosn

}
