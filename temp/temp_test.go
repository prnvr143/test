package temp

import (
	"fmt"
	"testing"
	"time"
)

func TestAnything(t *testing.T) {
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")
	fmt.Println(formattedTime)

	fmt.Println()

}
