package validator

import (
	"strconv"
)

// checks if password has at least 1 number
func hasNumber(p string) bool {
	for _, c := range p {
		_, err := strconv.ParseInt(string(c), 10, strconv.IntSize)
		if err == nil {
			return true
		}
	}
	return false
}

func passwordLongEnough(p string) bool {
	if 8 <= len(p) && len(p) <= 32 {
		return true
	}
	return false
}