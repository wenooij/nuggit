// Package keys defines utilities for working with nuggit FieldKeys.
//
// See nuggit.FieldKey.
package keys

import (
	"strconv"
	"strings"
)

func Cut(k string) (head, tail string, leaf bool) {
	head, tail, found := strings.Cut(k, ".")
	return head, tail, !found
}

func Split(k string) []string {
	return strings.Split(k, ".")
}

func Leaf(k string) bool {
	return !strings.ContainsRune(k, '.')
}

func Index(k string) (int64, bool) {
	if len(k) > 0 && '0' <= k[0] && k[0] <= '9' {
		if i, err := strconv.ParseInt(k, 10, 64); err == nil {
			return i, true
		}
	}
	return 0, false
}
