package utils

import (
	"fmt"
	"regexp"
	"strconv"
)

// check if password is strong
func IsStrong(p string) bool {
	re, err := regexp.Compile("\\d")
	if err != nil {
		panic(err)
	}
	return re.MatchString(p) && checkLength(p)
}

func checkLength(p string) bool {
	if 8 <= len(p) && len(p) <= 18 {
		return true
	}
	return false
}

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
