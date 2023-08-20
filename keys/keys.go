package keys

import (
	"strconv"
	"strings"
)

func Cut(k string) (head, tail string) {
	head, tail, _ = strings.Cut(string(k), ".")
	return
}

func Split(k string) []string {
	return strings.Split(string(k), ".")
}

func Index(k string) (int64, bool) {
	if len(k) > 0 && '0' <= k[0] && k[0] <= '9' {
		if i, err := strconv.ParseInt(string(k), 10, 64); err == nil {
			return i, true
		}
	}
	return 0, false
}
