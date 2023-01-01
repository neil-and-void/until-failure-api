package validator

import "regexp"

// checks if password has at least 1 number
func isComplex(p string) bool {
	re, err := regexp.Compile("\\d")
	if err != nil {
		panic(err)
	}
	return re.MatchString(p)
}

func passwordLongEnough(p string) bool {
	if 8 <= len(p) && len(p) <= 32 {
		return true
	}
	return false
}