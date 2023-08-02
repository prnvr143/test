package temp

// type transaction struct {
// 	From  string
// 	To    string
// 	Value int
// }

// func Start() {
// 	tx := transaction{
// 		From:  "add1",
// 		To:    "add2",
// 		Value: 100,
// 	}

// 	reqBodyBytes := new(bytes.Buffer)
// 	json.NewEncoder(reqBodyBytes).Encode(tx)

// 	by := reqBodyBytes.Bytes() // this is the []byte

// 	// tx.Encode()

// }

// // func (*transaction) Encode() string {
// // 	return base64.StdEncoding.Encode(*transaction)
// // }

// func Decode(s string) []byte {
// 	data, err := base64.StdEncoding.DecodeString(s)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return data
// }

// func Send(to string, from string, value int) string {
// 	t := &transaction{
// 		From:  from,
// 		To:    to,
// 		Value: value,
// 	}
// 	out, err := json.Marshal(t)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return string(out)

// }
