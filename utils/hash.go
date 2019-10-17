// Copyright (c) 2019.
// Author: Quan TRAN

package utils

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
)

// Hash returns SHA1 sum of input
func Hash(in interface{}) string {
	jsonBytes, _ := json.Marshal(in)
	return fmt.Sprintf("%x", sha1.Sum(jsonBytes))
}

// Hash returns SHA1 sum of input
func HashBytes(in []byte) string {
	return fmt.Sprintf("%x", sha1.Sum(in))
}
