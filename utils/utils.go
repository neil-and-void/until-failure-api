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
func GenerateVerificationCode(length int) (string, error) {
	// Generate a random byte slice of the specified length
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode the random byte slice using base64.URLEncoding, which produces a URL-safe string
	return base64.URLEncoding.EncodeToString(randomBytes), nil
}
