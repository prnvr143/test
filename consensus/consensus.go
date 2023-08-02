package consensus

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"sort"
	"strconv"

	"jumbochain.org/filemanagement"
	tx "jumbochain.org/transaction"
)

var leaderNode ValidatorInfo
var IsInValidatorPool bool = false

// number of concent neaded to form a block
// this is hardcoded, but will be percentage based
var NumberOfConcentToFormBlock int = 2

var concensus_aggrement int = 0

var TotalScore float64 = 0

var Validator_peerlist [][]string

type ValidatorInfo struct {
	Address   string  `json:"address"`
	MultiAddr string  `json:"multiAddrs"`
	Score     float64 `json:"score"`
}

type ByScore []ValidatorInfo

func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool { return a[i].Score > a[j].Score }

func topValidators(validatorPool [][]string, node_multiAddr string) (ValidatorInfo, []string, bool, bool, []ValidatorInfo) {

	isValidator := false
	isLeader := false
	var allValidatoInfo []ValidatorInfo
	for i := 0; i < len(validatorPool); i++ {
		validator_ := validatorPool[i]

		var validatoInfo ValidatorInfo

		validatoInfo.Address = validator_[0]
		validatoInfo.MultiAddr = validator_[1]

		float_value, _ := strconv.ParseFloat(validator_[2], 64)

		validatoInfo.Score = (float_value)

		// ////fmt.Println(validatoInfo)
		allValidatoInfo = append(allValidatoInfo, validatoInfo)

	}

	// sort the entries based on age in descending order
	sort.Sort(ByScore(allValidatoInfo))

	// slice the first 29 entries and store them in a separate variable
	//here set a percentage for total number of member in pool
	//currently I am including all the member in the pool
	numberOfValidators := len(validatorPool)
	top29 := allValidatoInfo[:numberOfValidators]
	var top29_string []string
	for i := 0; i < len(top29); i++ {
		validator_ := top29[i]

		if validator_.MultiAddr == node_multiAddr {
			isValidator = true
		}
		// validatorInfo_bytes, err := json.MarshalIndent(validator_, "", "")
		validatorInfo_bytes, err := json.Marshal(validator_)
		if err != nil {
			//fmt.Println(err)
		}
		top29_string = append(top29_string, string(validatorInfo_bytes))
	}

	top1 := allValidatoInfo[:1]

	if top1[0].MultiAddr == node_multiAddr {
		isLeader = true
	}

	return top1[0], top29_string, isValidator, isLeader, top29
}

func SelectValidators(node_multiAddr string) (bool, bool) {

	validatorpool := filemanagement.GetAllRecordsPeerlist("ValidatorPool.csv")
	leader, top29_string, isValidator, isLeader, _ := topValidators(validatorpool, node_multiAddr)
	updateLeader(leader)
	replaceValidators(top29_string)
	////fmt.Println("leader: ", leaderNode, isValidator, isLeader)

	return isValidator, isLeader

	// ////fmt.Println(leader, top29)
}

func updateLeader(leaderInfo ValidatorInfo) {
	leaderNode = leaderInfo
}

func replaceValidators(validators []string) {
	filemanagement.TruncateFile("Validators.csv")

	for i := 0; i < len(validators); i++ {
		filemanagement.AppendTofile("Validators.csv", validators[i])
	}

}

func FetchVerifiedTransactions() ([][]string, error) {

	//fetch transactions from TrxMemoryPoolValidator store it in varaible

	file, err := os.Open("VerifiedTransactions.csv")
	if err != nil {
		//fmt.Println("Error:", err)
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	transactions, err := reader.ReadAll()
	if err != nil {
		//fmt.Println("Error:", err)
		return nil, err
	}

	//empty csv file

	if err := os.Truncate("VerifiedTransactions.csv", 0); err != nil {
		//log.Printf("Failed to truncate: %v", err)
	}

	////fmt.Println(transactions)

	return transactions, nil

}

func AddVerificationConcent() {
	concensus_aggrement = concensus_aggrement + 1
}

func CheckVerificationConcent() int {
	return concensus_aggrement
}

func FetchTransactionHashes(transaction_bodies [][]string) []string {

	var transaction_hashes []string
	for i := 0; i < len(transaction_bodies); i++ {
		transaction_body := transaction_bodies[i][0]
		var tx_body tx.Transaction
		err := json.Unmarshal([]byte(transaction_body), &tx_body)
		if err != nil {
			////fmt.Println("json.Unmarshal([]byte(value)1, &transaction", err)
		}

		transaction_hashes = append(transaction_hashes, tx_body.Hash)

	}

	return transaction_hashes
}

func FetchTransactions(transaction_bodies [][]string) []string {

	var transactions []string
	for i := 0; i < len(transaction_bodies); i++ {
		transaction_body := transaction_bodies[i][0]
		transactions = append(transactions, transaction_body)

	}

	return transactions
}

// IMPORTANT: This will not work wehn one of the validator changes some data in a transaction, maybe.
// muje khud yaad nahi kyu likha mene ye, but Soch!
func VerifyTransaactionFromLeader(transactions_fromLeader []string) bool {

	// //this is fetching all trx from this node's memory pool and deleting it
	// transactons_ofValidaotr_, _ := FetchVerifiedTransactions()
	// transactons_ofValidaotr := FetchTransactions(transactons_ofValidaotr_)

	// //collect indexes of all the thansactions that matches with already verified transctions Of Leader
	// var indexesOfVMatchedTransactions_leader []int

	// //collect indexes of all the thansactions that does not matches with already verified transctions Of Validator
	// var indexesOfMatchedTransactions_validator []int

	// for i := 0; i < len(transactions_fromLeader); i++ {
	// 	transaction_fromLeader := transactions_fromLeader[i]

	// 	for j := 0; j < len(transactons_ofValidaotr); j++ {
	// 		transaction_ofValidator := transactons_ofValidaotr[j]

	// 		if transaction_ofValidator == transaction_fromLeader {
	// 			indexesOfVMatchedTransactions_leader = append(indexesOfVMatchedTransactions_leader, i)
	// 			indexesOfMatchedTransactions_validator = append(indexesOfMatchedTransactions_validator, j)
	// 			break
	// 		}

	// 	}
	// }

	//NO - comments for the above code, firgure it out!

	//it has all the verified transactions of leader

	//loop around
	// create transaction body
	// fetch hash

	//loop around our own verified transaction
	//fetch hash

	//also match status
	//if hash matched remove from our csv
	//if dosent matches, verify that body sepertly

	//set status of transaction true or false

	//fetch all verified transactions of this node

	//eveything from leader node should be check in our own verified transactions, those exsited should be removed from csv and those not existed should be verified and send concensus.
	//mil gaya answer?

	return true

}

func SetValidatorInfo(totalScore_ float64, isInValidatorpool_ bool) {
	TotalScore = totalScore_
	IsInValidatorPool = isInValidatorpool_

}

func SeperatePeerAndAdddressOfValidator() ([]string, []string) {
	validatorsInfo := Validator_peerlist

	var peers_validator []string
	var addresses_validator []string

	for i := len(validatorsInfo) - 1; i > -1; i-- {
		validator_info := validatorsInfo[i][0]
		var validatorInfo ValidatorInfo
		json.Unmarshal([]byte(validator_info), &validatorInfo)
		peers_validator = append(peers_validator, validatorInfo.MultiAddr)
		addresses_validator = append(addresses_validator, validatorInfo.Address)

	}

	return peers_validator, addresses_validator
}
