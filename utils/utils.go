package utils

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"strconv"
)

func StringToUInt(s string) uint {
	num, err := strconv.ParseUint(s, 10, strconv.IntSize)
	if err != nil {
		panic(err)
	}
	return uint(num)
}

func UIntToString(num uint) string {
	return fmt.Sprintf("%d", num)
}

// generate URL safe code
func GenerateVerificationCode(n int) (string, error) {
	b, err := GenerateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}
