package utils

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestRandBytes(t *testing.T) {
	data, err := RandBytes(8)
	fmt.Println(err)
	fmt.Println(len(data))
	fmt.Println(data)
	fmt.Println(hex.EncodeToString(data))
}
