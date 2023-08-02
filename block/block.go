package block

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"jumbochain.org/enum"
	"jumbochain.org/filemanagement"
	db "jumbochain.org/ldb"
	tr "jumbochain.org/transaction"
	"jumbochain.org/types"
)

// Define the structure of a block
type Block struct {
	BlockNumber        int      `json:"blocknumber"`
	Timestamp          string   `json:"timestamp"`
	Transactions       []string `json:"transactions"`
	PrevHash           string   `json:"prevHash"`
	Hash               string   `json:"hash"`
	ValidatorAddresses []string `json:"validatorAddresses"`
}

type BlockExtended struct {
	BlockNumber          int      `json:"blocknumber"`
	InvolvedTransactions []string `json:"involvedTransactions"`
	InvolvedPeers        []string `json:"involvedPeers"`
}

func BlockHash(b *Block) types.Hash {
	txInByte := []byte(fmt.Sprintf("%d-%s-%s", b.BlockNumber, b.Transactions, b.PrevHash))
	hash := sha256.Sum256(txInByte)
	return hash
}

func GenesisBlockHash(transactions []string) (string, string) {

	prevHash := "0000000000000000000000000000000000000000000000000000000000000000"
	timestamp := "123654789"
	block := &Block{
		BlockNumber:  0,
		Timestamp:    timestamp,
		Transactions: transactions,
		PrevHash:     prevHash,
	}

	blockHash := BlockHash(block)
	blockHash_string := blockHash.String()

	blockAfterHash := &Block{
		BlockNumber:  0,
		Timestamp:    timestamp,
		Transactions: transactions,
		PrevHash:     prevHash,
		Hash:         blockHash_string,
	}
	blockBody, _ := json.MarshalIndent(blockAfterHash, "", " ")

	fmt.Println(string(blockBody))

	return blockHash_string, string(blockBody)

}
func GetBalance(address string) int {

	balanceInBytes := db.FeatchFromDatabase("database", []byte(address))

	balance, _ := strconv.Atoi(string(balanceInBytes))

	return balance
}

func GetCurrentBlockNumber() string {
	currentBlockNumber := db.FeatchFromDatabase("database", []byte("currentBlockNumber"))

	return string(currentBlockNumber)
}

func GetBlockHashByNumber(blockNumber string) string {
	blockHash := db.FeatchFromDatabase("database", []byte(blockNumber))

	return string(blockHash)
}
func GetBlockByHash(blockNumber string) string {
	blockHash := db.FeatchFromDatabase("database", []byte(blockNumber))

	return string(blockHash)
}

func GetTransactionByHash(hashq string) string {
	transaction := db.FeatchFromDatabase("database", []byte(hashq))

	return string(transaction)
}

func CreateBlock(validator_addresses []string) ([]string, int) {
	transactions := filemanagement.GetAllRecordsMempool("VerifiedTransactions.csv")
	////fmt.Println("---=====")
	////fmt.Println(transactions)

	var transaction_hashes []string

	var involvedTransactions []string

	for i := 0; i < len(transactions); i++ {
		////fmt.Println(transactions[i][0])
		transaction := transactions[i][0]

		////fmt.Println("0000000")
		////fmt.Println(transaction)

		transactionBody, err := json.MarshalIndent(transaction, "", " ")
		if err != nil {
			//fmt.Println(err)
		}

		////fmt.Println(string(transactionBody))

		var tx tr.Transaction
		json.Unmarshal([]byte(transaction), &tx)

		////fmt.Println(tx.Hash)

		//add Transaction
		db.AddDataToDatabase("database", []byte(tx.Hash), []byte(transactionBody))

		from := tx.From
		to := tx.To
		value := tx.Value

		involvedTransactions = append(involvedTransactions, from)
		involvedTransactions = append(involvedTransactions, to)

		from_balance_string := string(db.FeatchFromDatabase("database", []byte(from)))
		to_balance_string := string(db.FeatchFromDatabase("database", []byte(to)))

		// strconv.ParseInt(from_balance_string)

		from_balance, err := strconv.ParseInt(from_balance_string, 10, 64)

		to_balance, err := strconv.ParseInt(to_balance_string, 10, 64)

		from_balance_i := int(from_balance)
		to_balance_i := int(to_balance)

		from_new_balance := from_balance_i - value
		to_new_balance := to_balance_i + value

		db.AddDataToDatabase("database", []byte(from), []byte(strconv.Itoa(from_new_balance)))
		db.AddDataToDatabase("database", []byte(to), []byte(strconv.Itoa(to_new_balance)))

		//update blockstate
		// db.AddDataToDatabase("database", []byte(blockNumber), []byte(hash))

		transaction_hashes = append(transaction_hashes, tx.Hash)

		// content, err := json.Marshal(txAfterHash)

	}

	////fmt.Println(transaction_hashes)

	prevHash := string(db.FeatchFromDatabase("database", []byte("lh")))
	lastBlock_string := string(db.FeatchFromDatabase("database", []byte("currentBlockNumber")))

	lastBlock, err := strconv.ParseInt(lastBlock_string, 10, 64)
	if err != nil {
		////fmt.Println(err)
	}

	createBlock(prevHash, transaction_hashes, int(lastBlock+1), validator_addresses)

	if err := os.Truncate("VerifiedTransactions.csv", 0); err != nil {
		////log.Printf("Failed to truncate: %v", err)
	}

	return involvedTransactions, int(lastBlock + 1)
}

func createBlock(prevHash string, transactions []string, blockNumber int, validator_addresses []string) {
	block := &Block{
		BlockNumber:        blockNumber,
		Timestamp:          time.Now().String(),
		Transactions:       transactions,
		PrevHash:           prevHash,
		ValidatorAddresses: validator_addresses,
	}
	BlockInBytes := []byte(fmt.Sprintf("%v", block))
	blockHash := types.HashFromBytes(BlockInBytes)
	blockHash_string := blockHash.String()

	blockAfterHash := &Block{
		BlockNumber:        blockNumber,
		Timestamp:          time.Now().String(),
		Transactions:       transactions,
		PrevHash:           prevHash,
		Hash:               blockHash_string,
		ValidatorAddresses: validator_addresses,
	}

	////fmt.Println(blockAfterHash)

	blockBody, err := json.MarshalIndent(blockAfterHash, "", " ")

	//fmt.Println(string(blockBody))

	if err != nil {
		////fmt.Println(err)
	}

	//update blockchain
	db.AddDataToDatabase("database", []byte(blockHash_string), []byte(blockBody))

	//update block number
	db.AddDataToDatabase("database", []byte(strconv.Itoa(blockNumber)), []byte(blockHash_string))

	lastHash := "lh"
	//update lasthash
	db.AddDataToDatabase("database", []byte(lastHash), []byte(blockHash_string))

	currentBlockNumber := "currentBlockNumber"
	//update current block number
	db.AddDataToDatabase("database", []byte(currentBlockNumber), []byte(strconv.Itoa(blockNumber)))

	//fmt.Println("Block created")
	//fmt.Println("current block number is :", strconv.Itoa(blockNumber))

}

//block code for node syncing is below

func JsonToBlock(jsonString string) ([]byte, Block) {
	blockBody, err := json.MarshalIndent(jsonString, "", " ")
	if err != nil {
		////fmt.Println(err)
	}

	////fmt.Println(string(blockBody))

	var block Block
	json.Unmarshal([]byte(jsonString), &block)

	////fmt.Println(block.BlockNumber)

	return blockBody, block
}

func AddlatestBlock(block_bytes []byte, block Block) {
	blockBody, err := json.MarshalIndent(block, "", " ")

	// ////fmt.Println(string(blockBody))

	if err != nil {
		////fmt.Println(err)
	}

	// update blockchain
	db.AddDataToDatabase("database", []byte(block.Hash), []byte(blockBody))

	//update block number
	db.AddDataToDatabase("database", []byte(strconv.Itoa(block.BlockNumber)), []byte(block.Hash))

	lastHash := "lh"
	//update lasthash
	db.AddDataToDatabase("database", []byte(lastHash), []byte(block.Hash))

	currentBlockNumber := "currentBlockNumber"
	//update current block number
	db.AddDataToDatabase("database", []byte(currentBlockNumber), []byte(strconv.Itoa(block.BlockNumber)))

}

// Create a new block with a previous hash
// func Createblock(index int, transactions []string, prevHash string) *Block {
// 	block := &Block{
// 		Index:        index,
// 		Timestamp:    time.Now().String(),
// 		Transactions: transactions,
// 		PrevHash:     prevHash,
// 	}
// 	block.Hash = calculateHash(block)
// 	return block
// }

// func AddDataToDatabase(k []byte, v []byte) {
// 	db, err := leveldb.OpenFile("database", nil)
// 	if err != nil {
// 		panic(err)
// 	}
// 	db.Put(k, v, nil)
// }

// func FeatchFromDatabase(abc []byte) []byte {
// 	db, _ := leveldb.OpenFile("database", nil)
// 	data, _ := db.Get(abc, nil)
// 	////fmt.Println("%v \n", data)

// }

// func MakeGenesisBlock() {
// 	utilsP.ReadFromTxData()

// 	// Create the genesis block
// 	genesisblock := &block{
// 		Index:     0,
// 		Timestamp: time.Now().String(),
// 		Data:      "Genesis block",
// 		PrevHash:  "",
// 	}
// 	genesisblock.Hash = calculateHash(genesisblock)
// 	blockchain = append(blockchain, genesisblock)
// 	blockchainBytes, err := json.MarshalIndent(blockchain, "", "    ")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	////fmt.Println(string(blockchainBytes))
// 	blockNumber := []byte("0")
// 	AddDataToDatabase(blockNumber, blockchainBytes)
// }

// func AddBlockToChain() {
// 	genesisBlockNumber := []byte("0")
// 	genesisBlock := FeatchFromDatabase(genesisBlockNumber)
// 	// if genesisBlock != 0 {
// 	// 	////fmt.Println("genesis is not present")
// 	// }
// 	////fmt.Println(genesisBlock)

// 	// Add some transactions to the blockchain
// 	data := utils.ReadFromTxData()
// 	addblock(data)
// 	//addblock("Transaction 2")

// 	// Print the blockchain
// 	blockchainBytes, err := json.MarshalIndent(blockchain, "", "    ")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	////fmt.Println(string(blockchainBytes))
// }

// Calculate the hash of a block
// func calculateHash(block *Block) string {
// 	record := strconv.Itoa(block.Index) + block.Timestamp + block.Transactions[0] + block.PrevHash
// 	h := sha256.New()
// 	h.Write([]byte(record))
// 	hash := h.Sum(nil)
// 	return hex.EncodeToString(hash)
// }

// // // Define the blockchain
// var blockchain []*Block

// // Add a new block to the blockchain
//func addblock(data string) {
// prevblock := blockchain[len(blockchain)-1]
// newblock := createblock(prevblock.Index+1, data, prevblock.Hash)
//blockchain = append(blockchain, newblock)

// func addblock(data ...string) {
// 	prevblock := blockchain[len(blockchain)-1]
// 	newData := ""
// 	for _, d := range data {
// 		newData += d
// 	}
// 	newblock := Createblock(prevblock.Index+1, newData, prevblock.Hash)
// 	blockchain = append(blockchain, newblock)
// }

// // func main() {
// // 	// Create the genesis block
// // 	genesisblock := &block{
// // 		Index:     0,
// // 		Timestamp: time.Now().String(),
// // 		Data:      "Genesis block",
// // 		PrevHash:  "",
// // 	}
// // 	genesisblock.Hash = calculateHash(genesisblock)
// // 	blockchain = append(blockchain, genesisblock)

// // 	// Add some transactions to the blockchain
// // 	addblock("Transaction1", "Transaction2", "Transaction3", "myTransaction")
// // 	//addblock("Transaction 2")

// // 	// Print the blockchain
// // 	blockchainBytes, err := json.MarshalIndent(blockchain, "", "    ")
// // 	if err != nil {
// // 		log.Fatal(err)
// // 	}
// // 	////fmt.Println(string(blockchainBytes))
// // }

func GetBlockByString(block_str string) Block {
	var block Block
	json.Unmarshal([]byte(block_str), &block)

	return block

}

func CreateBlockExtended(involvedTransactions []string, involvedPeers []string, blockNumber int) {

	blockExtended := &BlockExtended{
		BlockNumber:          blockNumber,
		InvolvedTransactions: involvedTransactions,
		InvolvedPeers:        involvedPeers,
	}

	blockExtendedBody, _ := json.MarshalIndent(blockExtended, "", " ")
	db.AddDataToDatabase(string(enum.Blockextended), []byte(strconv.Itoa(blockNumber)), (blockExtendedBody))

	//update current block extended number
	currentBlockExtendedNumber := string(enum.CurrentBlockExtendedNumber)

	db.AddDataToDatabase(string(enum.Blockextended), []byte(currentBlockExtendedNumber), []byte(strconv.Itoa(blockNumber)))
}

func AddLatestBlockExtended(blockExtended_string string) {
	var blockExtended BlockExtended
	json.Unmarshal([]byte(blockExtended_string), &blockExtended)

	db.AddDataToDatabase(string(enum.Blockextended), []byte(strconv.Itoa(blockExtended.BlockNumber)), []byte(blockExtended_string))

	currentBlockExtendedNumber := string(enum.CurrentBlockExtendedNumber)

	db.AddDataToDatabase(string(enum.Blockextended), []byte(currentBlockExtendedNumber), []byte(strconv.Itoa(blockExtended.BlockNumber)))

}

func GetBlockExtendedByNumber(blockNumber string) string {
	blockExtended := db.FeatchFromDatabase(string(enum.Blockextended), []byte(blockNumber))

	return string(blockExtended)
}

func GetCurrentBlockExtendedNumber() string {
	currentBlockNumber := db.FeatchFromDatabase(string(enum.Blockextended), []byte(string(enum.CurrentBlockExtendedNumber)))

	return string(currentBlockNumber)
}

func GetExecutedBlockExtendedNumber() string {
	executedBlockNumber := db.FeatchFromDatabase(string(enum.Blockextended), []byte(string(enum.ExecutedBlockExtendednumber)))

	return string(executedBlockNumber)
}

func GetBlockExtendedInfo(blockExtendedNumber int) ([]string, []string) {
	blockExtendedNumber_str := strconv.Itoa(blockExtendedNumber)
	blockExtendedBody_str := GetBlockExtendedByNumber(blockExtendedNumber_str)

	var blockExtended BlockExtended
	json.Unmarshal([]byte(blockExtendedBody_str), &blockExtended)

	return blockExtended.InvolvedTransactions, blockExtended.InvolvedPeers

}

func DeleteBlockExtendedData(blockExtendedNumber int) {
	blockExtendedNumber_str := strconv.Itoa(blockExtendedNumber)
	db.DeleteFromDatabase(string(enum.Blockextended), []byte(blockExtendedNumber_str))
}

func GetBlockExtendedByString(blockExtended_str string) {

}

func GetTransactionsOfBlockNumber(blockExtendedNumber_str string) []string {
	blockHash := GetBlockHashByNumber(blockExtendedNumber_str)
	blockDetails := GetBlockByHash(blockHash)
	block := GetBlockByString(blockDetails)

	transactionHashes := block.Transactions

	var transactionbodies []string
	for i := 0; i < len(transactionHashes); i++ {
		transactionBody_string := GetTransactionByHash(transactionHashes[i])
		transactionbodies = append(transactionbodies, transactionBody_string)
	}

	return transactionbodies
}

func SaveTransactionsInLDB(transactionbodies []string) {

	for i := 0; i < len(transactionbodies); i++ {
		transactionBody_str := transactionbodies[i]

		var transaction tr.Transaction
		json.Unmarshal([]byte(transactionBody_str), &transaction)

		hash := transaction.Hash

		db.AddDataToDatabase(string(enum.Database), []byte(hash), []byte(transactionBody_str))

	}
}
