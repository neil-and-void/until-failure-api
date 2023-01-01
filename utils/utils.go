package utils

import (
	"fmt"
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
