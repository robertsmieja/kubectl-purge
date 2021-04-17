package util

import "strings"

func Contains(arr []string, str string) bool {
	for _, item := range arr {
		if str == item {
			return true
		}
	}
	return false
}

func StartsWithAny(prefixArr []string, str string) bool {
	for _, prefix := range prefixArr {
		if strings.HasPrefix(str, prefix) {
			return true
		}
	}
	return false
}
