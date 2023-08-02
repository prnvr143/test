package transaction

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"jumbochain.org/accounts"
	"jumbochain.org/filemanagement"
	ldb "jumbochain.org/ldb"
	"jumbochain.org/temp"
	"jumbochain.org/types"
)

var verifiedTransactions []Transaction

type Transaction struct {
	To         string              `json:"to"`
	From       string              `json:"from"`
	Value      int                 `json:"value"`
	PubKeyFrom string              `json:"fromPublicKey"`
	Hash       string              `json:"hashInString"`
	Signature  *accounts.Signature `json:"signature"`
	Data       []byte              `json:"data"`
	Timestamp  int64
}

func SignGenesisTrx(_to string, _value int) (string, string) {
	var timestamp int64 = 123654789
	_from := "dh00000000000000000GENESIS0000000000000000"
	tx := &Transaction{
		To:        _to,
		From:      _from,
		Value:     _value,
		Timestamp: timestamp,
	}
	transactionInBytes := []byte(fmt.Sprintf("%v", tx))
	txHash := types.HashFromBytes(transactionInBytes)
	txxHash := txHash.String()

	txAfterHash := &Transaction{
		To:    _to,
		From:  _from,
		Value: _value,
		Hash:  txxHash,
	}

	// var trx Transaction
	transactionBody, err := json.MarshalIndent(txAfterHash, "", " ")

	////fmt.Println(string(transactionBody))

	// content, err := json.Marshal(txAfterHash)
	if err != nil {
		////fmt.Println(err)
	}
	// ////fmt.Println(content)
	// ////fmt.Println(txxHash)

	// var trxBody Transaction

	// json.Unmarshal(transactionBody, &trxBody)
	// ////fmt.Println(trxBody)

	return txxHash, string(transactionBody)

	//

	// return txHash
}
func TransactionHash(tx *Transaction) types.Hash {
	txInByte := []byte(fmt.Sprintf("%s-%s-%d", tx.From, tx.To, tx.Value))
	hash := sha256.Sum256(txInByte)
	return hash
}

func SendTxx2(from string, to string, value int, auth string) types.Hash {
	timestamp := time.Now().Unix()
	tx := &Transaction{
		To:        to,
		From:      from,
		Value:     value,
		Timestamp: timestamp,
	}
	hash := TransactionHash(tx)                                      // calculateing transaction hash
	fromPrivatekey := accounts.GetPrivateKeyFromKeystore(from, auth) // getting private key from keystore
	fromPublicKey := &fromPrivatekey.PublicKey
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(fromPublicKey) // converting public key to string
	if err != nil {
		panic(err)
	}
	publicKeyEncoded := base64.StdEncoding.EncodeToString(publicKeyBytes)
	publicKeyBytes, err = base64.StdEncoding.DecodeString(publicKeyEncoded)
	if err != nil {
		panic(err)
	}
	pubKey, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		panic(err)
	}
	r, s, err := ecdsa.Sign(rand.Reader, fromPrivatekey, hash[:]) // signing transaction
	if err != nil {
		//fmt.Println("Error signing transaction", err)
	}
	ver := ecdsa.Verify(pubKey.(*ecdsa.PublicKey), hash[:], r, s) // verifying transaction
	fmt.Println("verify transaction", ver)
	sign := accounts.Signature{
		S: s,
		R: r,
	}
	txAfterHash := &Transaction{
		To:         to,
		From:       from,
		PubKeyFrom: publicKeyEncoded,
		Value:      value,
		Hash:       hash.String(),
		Signature:  &sign,
	}
	notHavingFile := checkFile("TrxMemPool.csv")
	if notHavingFile != nil {
		//fmt.Println("Unable to create file", notHavingFile)
	}
	f, err := os.OpenFile("TrxMemPool.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		//fmt.Println(err)
	}
	defer f.Close()
	content, err := json.Marshal(txAfterHash)
	if err != nil {
		//fmt.Println(err)
	}
	n, err := f.Write(content)
	if err != nil {
		fmt.Println(n, err)
	}
	if n, err = f.WriteString("\n"); err != nil {
		//fmt.Println(n, err)
	}
	return hash
}

func MakeGenesis(_to string, _from string, _value int) (types.Hash, error) {
	var timestamp int64 = 123654789

	tx := &Transaction{
		To:        _to,
		From:      _from,
		Value:     _value,
		Timestamp: timestamp,
	}

	txHash := TransactionHash(tx)
	notHavingFile := checkFile("txData.csv")
	if notHavingFile != nil {
		////fmt.Println("Unable to create file", notHavingFile)
	}
	////fmt.Println("TRX:-", txHash)

	return txHash, nil
}

func checkFile(filename string) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		_, err := os.Create(filename)
		if err != nil {
			return err
		}
	}
	return nil
}

func Temp_VerifyTransaction() {
	nonVerifiedMempool := filemanagement.GetAllRecordsMempool("TrxMemPoolValidator.csv")
	filemanagement.TruncateFile("TrxMemPoolValidator.csv")
	for i := 0; i < len(nonVerifiedMempool); i++ {
		filemanagement.AppendTofile("VerifiedTransactions.csv", nonVerifiedMempool[i][0])
	}
}

//this function is not working properly, balance verification is not working!
//for now I have used above temporary function
//test this one, and re write it is more readable way

//shiv

var hold string

func VerifyTransactions() (bool, error) { // used to verify the transactions on validator node
	var transaction Transaction
	csvData := temp.ReadCsv("TrxMemPoolValidator.csv")
	CSVMap := make(map[int]Transaction, 0)
	VerifiedTransactionsMap := make(map[int]Transaction, 0)
	////fmt.Println("==================", csvData)

	for i := 0; i < len(csvData); i++ {
		////fmt.Println("================", csvData[i][0])
		err := json.Unmarshal([]byte(csvData[i][0]), &transaction)
		if err != nil {
			////fmt.Println("json.Unmarshal([]byte(value)1, &transaction", err)
		}

		////fmt.Println(transaction.From)
		////fmt.Println("Public Key ===========>", transaction.PubKeyFrom)
		publicKeyEncoded := transaction.PubKeyFrom
		publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyEncoded)
		if err != nil {
			panic(err)
		}
		publicKey, err := x509.ParsePKIXPublicKey(publicKeyBytes)
		if err != nil {
			panic(err)
		}
		if err != nil {
			panic(err)
		}
		tx := &Transaction{
			To:        transaction.To,
			From:      transaction.From,
			Value:     transaction.Value,
			Timestamp: transaction.Timestamp,
		}
		hash := TransactionHash(tx)
		verification := ecdsa.Verify(publicKey.(*ecdsa.PublicKey), hash[:], transaction.Signature.R, transaction.Signature.S)
		////fmt.Println("verification:", verification)

		//Open the LevelDB file for reading
		fromValue, err := ldb.GetDataFromLevelDB(transaction.From)

		if err != nil {
			////fmt.Println("json.Unmarshal([]byte(value)2, &transaction", err)
		}

		if transaction.From != "" && fromValue >= float64(transaction.Value) && verification == true {
			// delete(CSVMap, i)
			VerifiedTransactionsMap[i] = transaction
		} else {
			CSVMap[i] = transaction
			// delete(VerifiedTransactionsMap, i)
		}
		//transaction
	}
	temp.TruncateCSVFile()
	for _, value := range CSVMap {

		//////fmt.Println("csv file data ", CSVMap[i])
		//////fmt.Println("verified transaction data ", VerifiedTransactionsMap[i])
		marshData, err := json.Marshal(value)
		if err != nil {
			////fmt.Println("json.Unmarshal([]byte(value)1, &transaction", err)
		}
		temp.AddDataToCSVFile("TrxMemPoolValidator.csv", string(marshData))

	}
	for _, value := range VerifiedTransactionsMap {

		//////fmt.Println("csv file data ", CSVMap[i])
		//////fmt.Println("verified transaction data ", VerifiedTransactionsMap[i])
		marshData, err := json.Marshal(value)
		if err != nil {
			////fmt.Println("json.Unmarshal([]byte(value)1, &transaction", err)
		}
		temp.AddDataToCSVFile("VerifiedTransactions.csv", string(marshData))

	}

	return false, nil

}

//--------------------------

func StoreTransactionsInMemory(transactions []string) {

	for i := 0; i < len(transactions); i++ {
		transaction_string := transactions[i]

		var tx Transaction
		json.Unmarshal([]byte(transaction_string), &tx)

		verifiedTransactions = append(verifiedTransactions, tx)

	}
}

func SaveTransactionsInLDB() {

	var verifiedTransactions []Transaction = verifiedTransactions
	verifiedTransactions = []Transaction{}

	for i := 0; i < len(verifiedTransactions); i++ {
		tx_body := verifiedTransactions[i]
		transactionBody, err := json.MarshalIndent(tx_body, "", " ")
		if err != nil {
			//fmt.Println(err)
		}

		//add Transaction
		ldb.AddDataToDatabase("database", []byte(tx_body.Hash), []byte(transactionBody))

		from := tx_body.From
		to := tx_body.To
		value := tx_body.Value

		from_balance_string := string(ldb.FeatchFromDatabase("database", []byte(from)))
		to_balance_string := string(ldb.FeatchFromDatabase("database", []byte(to)))

		// strconv.ParseInt(from_balance_string)

		from_balance, err := strconv.ParseInt(from_balance_string, 10, 64)

		to_balance, err := strconv.ParseInt(to_balance_string, 10, 64)

		from_balance_i := int(from_balance)
		to_balance_i := int(to_balance)

		from_new_balance := from_balance_i - value
		to_new_balance := to_balance_i + value

		ldb.AddDataToDatabase("database", []byte(from), []byte(strconv.Itoa(from_new_balance)))
		ldb.AddDataToDatabase("database", []byte(to), []byte(strconv.Itoa(to_new_balance)))

	}

}

// func ProcessBlockExtended(involvedAddresses []string, involvedPeers []string, blockExtendedNumber int) {

// 	keystoreAddresses := filemanagement.GetAllRecords(string(enum.Keystore))

// 	m := make(map[string]bool)
// 	var commonAddresses []string

// 	// Store elements from the first column of arr1 in the map
// 	for _, row := range keystoreAddresses {
// 		m[row[0]] = true
// 	}

// 	// Check elements from arr2 against the map
// 	for _, str := range involvedAddresses {
// 		if m[str] {
// 			commonAddresses = append(commonAddresses, str)
// 		}
// 	}

// 	if len(commonAddresses) > 0 {

// 		for i := 0; i < len(involvedPeers); i++ {
// 			// target := involvedPeers[i]

// 		}

// 	}

// 	//delete blockExtendedNumber's information

// }
