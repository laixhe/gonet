package utils

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestRandBytes(t *testing.T) {
	data := RandBytes(1)
	fmt.Println(data)
	fmt.Println(hex.EncodeToString(data))
	data = RandBytes(2)
	fmt.Println(data)
	fmt.Println(hex.EncodeToString(data))
	data = RandBytes(3)
	fmt.Println(data)
	fmt.Println(hex.EncodeToString(data))
	data = RandBytes(4)
	fmt.Println(data)
	fmt.Println(string(data))
	fmt.Println(hex.EncodeToString(data))
}

func TestRandBool(t *testing.T) {
	fmt.Println(RandBool())
	fmt.Println(RandBool())
	fmt.Println(RandBool())
	fmt.Println(RandBool())
}

func TestRandRange(t *testing.T) {
	fmt.Println(RandRange(0, 5))
	fmt.Println(RandRange(0, 5))
	fmt.Println(RandRange(0, 5))
	fmt.Println(RandRange(0, 5))
	fmt.Println(RandRange(0, 5))
	fmt.Println(RandRange(0, 5))
}

func TestRandLetter(t *testing.T) {
	fmt.Println(RandLetter(5))
	fmt.Println(RandLetter(5))
	fmt.Println(RandLetter(5))
	fmt.Println(RandLetter(5))
	fmt.Println(RandLetter(5))
	fmt.Println(RandLetter(5))
	fmt.Println()
	fmt.Println(RandLetter(5, false))
	fmt.Println(RandLetter(5, false))
	fmt.Println(RandLetter(5, false))
	fmt.Println(RandLetter(5, false))
	fmt.Println(RandLetter(5, false))
	fmt.Println(RandLetter(5, false))
	fmt.Println()
	fmt.Println(RandLetter(5, true))
	fmt.Println(RandLetter(5, true))
	fmt.Println(RandLetter(5, true))
	fmt.Println(RandLetter(5, true))
	fmt.Println(RandLetter(5, true))
	fmt.Println(RandLetter(5, true))
}

func TestRandString(t *testing.T) {
	fmt.Println(RandString(5))
	fmt.Println(RandString(5))
	fmt.Println(RandString(5))
	fmt.Println(RandString(5))
	fmt.Println(RandString(5))
	fmt.Println()
	fmt.Println(RandString(5, false))
	fmt.Println(RandString(5, false))
	fmt.Println(RandString(5, false))
	fmt.Println(RandString(5, false))
	fmt.Println(RandString(5, false))
	fmt.Println()
	fmt.Println(RandString(5, true))
	fmt.Println(RandString(5, true))
	fmt.Println(RandString(5, true))
	fmt.Println(RandString(5, true))
	fmt.Println(RandString(5, true))
}

func TestRandNumeral(t *testing.T) {
	fmt.Println(RandNumeral(1))
	fmt.Println(RandNumeral(2))
	fmt.Println(RandNumeral(3))
	fmt.Println(RandNumeral(4))
	fmt.Println(RandNumeral(5))
}
