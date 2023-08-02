package filemanagement

import (
	"testing"
)

func TestAddingArray(t *testing.T) {
	data := []string{"address9", "peerid3", "score3"}
	AddArrayToFile("test.csv", data)

}

func TestRemoveArray(t *testing.T) {
	walletAddress := "address7"
	RemoveArrayFromFile("test.csv", walletAddress)
}
