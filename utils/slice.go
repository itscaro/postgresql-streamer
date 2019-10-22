// Copyright (c) 2019.
// Author: Quan TRAN

package utils

// Map applies a function on each item of the slice
func Map(list []string, f func(string) string) []string {
	result := make([]string, len(list))
	for i, item := range list {
		result[i] = f(item)
	}
	return result
}
