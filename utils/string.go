// Copyright (c) 2019.
// Author: Quan TRAN

package utils

import "regexp"

// SanitizeString replaces anything which is not alphanum with hyphen
func SanitizeString(str string) string {
	re := regexp.MustCompile("(?i)[^a-z0-9]")
	return re.ReplaceAllString(str, "-")
}
