package filemanagement

import (
	"encoding/csv"
	"os"
	"time"
)

func GetAllRecords(filename string) [][]string {
	f, _ := os.Open(filename)

	r := csv.NewReader(f)

	defer f.Close()

	records, _ := r.ReadAll()

	return records

}

func GetAllRecordsMempool(filename string) [][]string {
	f, err := os.Open(filename)

	if err != nil {
		////fmt.Println("error is opening file")
	}

	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ';'
	r.LazyQuotes = true

	records, err := r.ReadAll()

	if err != nil {
		////fmt.Println("error in reading file")
	}

	return records

}

func GetAllRecordsPeerlist(filename string) [][]string {
	f, err := os.Open(filename)

	if err != nil {
		////fmt.Println("error is opening file")
	}

	r := csv.NewReader(f)

	defer f.Close()

	records, err := r.ReadAll()

	if err != nil {
		////fmt.Println("error in reading file")
	}

	return records

}

func AppendTofile(filename string, data string) error {

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		////fmt.Println("Could not open example.txt")
		return err
	}

	defer file.Close()

	_, err2 := file.WriteString(data + "\n")

	if err2 != nil {
		////fmt.Println("Could not write text to example.txt")
		return err2

	} else {
		////fmt.Println("Operation successful! Text has been appended")
	}

	return nil

}

func GetNumberOfRecords(filename string) int {
	f, err := os.Open(filename)

	if err != nil {
		////fmt.Println("error is opening file")
	}

	defer f.Close()

	r := csv.NewReader(f)

	records, err := r.ReadAll()

	if err != nil {
		////fmt.Println("error in reading file")
	}

	return len(records)
}

func IsAlreadyInFile(filename string, multiaddr string) bool {
	f, err := os.Open(filename)

	if err != nil {
		////fmt.Println("error is opening file")
	}

	defer f.Close()

	r := csv.NewReader(f)

	records, err := r.ReadAll()

	if err != nil {
		////fmt.Println("error in reading file")
	}

	for i := 0; i < len(records); i++ {
		record := records[i][0]

		if record == multiaddr {
			return true
		}
	}

	return false

}

func IsAlreadyInValidatorPool(filename string, validaotrInfo []string) bool {
	f, err := os.Open(filename)

	if err != nil {
		////fmt.Println("error is opening file")
	}

	defer f.Close()

	r := csv.NewReader(f)

	records, err := r.ReadAll()

	if err != nil {
		////fmt.Println("error in reading file")
	}

	for i := 0; i < len(records); i++ {
		record := records[i]

		if record[0] == validaotrInfo[0] {
			return true
		}
	}

	return false

}

func TruncateFile(filename string) {
	if err := os.Truncate(filename, 0); err != nil {
		////log.Printf("Failed to truncate: %v", err)
	}
}

func AddArrayToFile(filename string, data []string) {
	isFileReady := false

	for isFileReady == false {
		//"temp/VerifiedTransactions.csv"
		// Adding array to CSV
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		// file, err := os.Create("example.csv")

		if err == nil {
			defer file.Close()

			writer := csv.NewWriter(file)
			defer writer.Flush()

			err = writer.Write(data)
			if err != nil {
				panic(err)
			}
			isFileReady = true
		} else {
			//fmt.Println("file is closed wait ")
			time.Sleep(1 * time.Second)
		}

	}

}

func RemoveArrayFromFile(filename string, walletAddress string) {
	isFileReady := false

	for isFileReady == false {
		f, err := os.Open(filename)

		if err == nil {
			r := csv.NewReader(f)

			defer f.Close()

			records, err := r.ReadAll()

			if err != nil {
				//fmt.Println("error in reading file")
			}

			TruncateFile(filename)

			for i := 0; i < len(records); i++ {
				array := records[i]
				if array[0] != walletAddress {
					AddArrayToFile(filename, array)
				}
			}
			isFileReady = true
		} else {
			//fmt.Println("file is closed wait ")
			time.Sleep(1 * time.Second)
		}
	}
}

func AddArrayToFile1(filename string, data []string) {
	//"temp/VerifiedTransactions.csv"
	// Adding array to CSV
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// file, err := os.Create("example.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write(data)
	if err != nil {
		panic(err)
	}
}

func RemoveArrayFromFile1(filename string, walletAddress string) {
	f, err := os.Open(filename)

	if err != nil {
		////fmt.Println("error is opening file")
	}
	r := csv.NewReader(f)

	defer f.Close()

	records, err := r.ReadAll()

	if err != nil {
		////fmt.Println("error in reading file")
	}

	TruncateFile(filename)

	for i := 0; i < len(records); i++ {
		array := records[i]
		if array[0] != walletAddress {
			AddArrayToFile(filename, array)
		}
	}

}
