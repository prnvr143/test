package agent

import (
	"encoding/json"
	"testing"

	"jumbochain.org/ldb"
)

func TestLDB(t *testing.T) {
	key := "hii"
	value_ := "hii_value"

	value := ldb.GetInfoDB("validatordb", []byte(key))

	if len(value) < 1 {
		ldb.AddInfoDB("validatordb", []byte(key), []byte(value_))
	}
}

func TestSystemRequirement(t *testing.T) {
	body_InBytes := FetchSystemRequirements("address", "peerid")

	var user_SystemInfo User_SystemInfo
	json.Unmarshal(body_InBytes, &user_SystemInfo)

	////fmt.Println(user_SystemInfo)
	////fmt.Println("")
}
