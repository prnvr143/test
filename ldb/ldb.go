package ldb

import (
	"encoding/json"
	"log"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

// @Vaikr
func AddDataToDatabase1(k []byte, v []byte) string {
	db, err := leveldb.OpenFile("savedKeys", nil)
	if err != nil {
		panic(err)
	}
	db.Put(k, v, nil)
	abc := "stored successfully"
	return abc
}

// @Vaikr
func FeatchFromDatabase1(abc []byte) ([]byte, error) {
	db, _ := leveldb.OpenFile("savedKeys", nil)
	data, _ := db.Get(abc, nil)
	//////fmt.Println("%v \n", data)
	return data, nil

}

func AddDataToDatabase(databasename string, key []byte, value []byte) {
	isDbReady := false

	for isDbReady == false {
		db, err := leveldb.OpenFile(databasename, nil)
		if err == nil {
			db.Put(key, value, nil)
			isDbReady = true
			db.Close()
		} else {
			////fmt.Println("db is closed wait ")
			time.Sleep(2 * time.Second)
		}
	}

}

func FeatchFromDatabase(databasename string, key []byte) []byte {
	isDbReady := false

	var data []byte
	for isDbReady == false {
		db, err := leveldb.OpenFile(databasename, nil)
		if err == nil {
			data_, _ := db.Get(key, nil)
			data = data_
			isDbReady = true
			db.Close()
		} else {
			////fmt.Println("db is closed wait ")
			time.Sleep(2 * time.Second)
		}
	}
	return data

}

func DeleteFromDatabase(databasename string, key []byte) {
	db, err := leveldb.OpenFile(databasename, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Perform the delete operation
	err = db.Delete(key, nil)
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println("Data deleted successfully!")
}

// func AddDataToDatabase(databasename string, key []byte, value []byte) {
// 	db, err := leveldb.OpenFile(databasename, nil)
// 	defer db.Close()
// 	if err != nil {
// 		////fmt.Println("errrorrrrrrr  AddDataToDatabase")
// 		panic(err)
// 	}
// 	db.Put(key, value, nil)
// }

// func FeatchFromDatabase(databasename string, key []byte) []byte {
// 	db, err := leveldb.OpenFile(databasename, nil)
// 	defer db.Close()

// 	if err != nil {
// 		////fmt.Println("errrorrrrrrr  FeatchFromDatabase  OpenFile")
// 		panic(err)
// 	}

// 	data, err2 := db.Get(key, nil)
// 	if err2 != nil {
// 		////fmt.Println("errrorrrrrrr  FeatchFromDatabase  Get")
// 		panic(err)
// 	}
// 	// ////fmt.Println("%v \n", data)

// 	return data

// }

func GetDataFromLevelDB(from string) (float64, error) {

	db, err := leveldb.OpenFile("database", nil)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	//time.Sleep(time.Second)

	////fmt.Println(from)

	////fmt.Println("-----------")

	fromBalanceBytes, err := db.Get([]byte(from), nil)

	////fmt.Println(fromBalanceBytes)

	////fmt.Println("-----------")

	//var transaction models.Transaction
	var amount float64
	err = json.Unmarshal(fromBalanceBytes, &amount)
	if err != nil {
		////fmt.Println("json.Unmarshal([]byte(value)3, &transaction", err)
	}
	////fmt.Println(amount)

	return amount, nil

}

func GetInfoDB(databasename string, key []byte) []byte {
	isDbReady := false

	var data []byte
	for isDbReady == false {
		db, err := leveldb.OpenFile(databasename, nil)
		if err == nil {
			data_, _ := db.Get(key, nil)
			data = data_
			isDbReady = true
			db.Close()
		} else {
			////fmt.Println("db is closed wait ")
			time.Sleep(2 * time.Second)
		}
	}
	return data

}

func AddInfoDB(databasename string, key []byte, value []byte) {
	isDbReady := false

	for isDbReady == false {
		db, err := leveldb.OpenFile(databasename, nil)
		if err == nil {
			db.Put(key, value, nil)
			isDbReady = true
			db.Close()
		} else {
			////fmt.Println("db is closed wait ")
			time.Sleep(2 * time.Second)
		}
	}
}

func GetInfoDB1(databasename string, key []byte) []byte {
	db, err := leveldb.OpenFile(databasename, nil)
	defer db.Close()
	if err == nil {
		data_, _ := db.Get(key, nil)
		return data_
	} else {
		panic(err)
	}

}

func AddInfoDB1(databasename string, key []byte, value []byte) {
	db, err := leveldb.OpenFile(databasename, nil)
	defer db.Close()
	if err == nil {
		db.Put(key, value, nil)

	} else {
		panic(err)
	}
}
